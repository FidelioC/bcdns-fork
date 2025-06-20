package verifying_network

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	// https://pkg.go.dev/github.com/libp2p/go-libp2p#section-readme

	"github.com/khalidzahra/dns_client/substrate"
	libp2p "github.com/libp2p/go-libp2p"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	host "github.com/libp2p/go-libp2p/core/host"
	peer "github.com/libp2p/go-libp2p/core/peer"
	peerstore "github.com/libp2p/go-libp2p/core/peerstore"
	ma "github.com/multiformats/go-multiaddr" //https://github.com/multiformats/multiaddr
)

// topic name should always be the same for all nodes in the verifying network
const TopicName = "verifying-network" 

type VerifierNode struct {
	host  host.Host
	ps    *pubsub.PubSub
	topic *pubsub.Topic
	sub   *pubsub.Subscription
}

func NewVerifierNode(ctx context.Context, bootstrap string) *VerifierNode {
	// create the libp2p node (host)
	newHost, err := libp2p.New() 
	if err != nil {
		log.Fatal(err)
	}

	// if bootstrap address was provided, to talk to other existing peers
	if bootstrap != "" {
		// parse the multiaddr string, e.g., /ip6/2604:3d09:a98d:b100:4552:be06:a3ca:b295/udp/64042/webrtc-direct/certhash/uEiAAKlDpHtl0D3aOBbJEEYqArQLTuZ9zL-smFMJ17JGrag/p2p/12D3KooWAVaoXdP8wurmgFXizqKHV4NGZnHzJxpKdmyAvfS9tEW1
		maddr, err := ma.NewMultiaddr(bootstrap)
		if err != nil {
			log.Fatal(err)
		}

		// convert the maddr to peer.Addrinfo, which separates the peer ID and addresses
		info, err := peer.AddrInfoFromP2pAddr(maddr)
		if err != nil {
			log.Fatal(err)
		}
		
		// add the address to this node's internal peer storage
		newHost.Peerstore().AddAddrs(info.ID, info.Addrs, peerstore.PermanentAddrTTL)

		// connect to peer
		newHost.Connect(ctx, *info)
	}
	// initialize new GossipSub
	pubSub, err := pubsub.NewGossipSub(ctx, newHost)
	if err != nil {
		log.Fatal(err)
	}

	// join on a pubsub topic (constant), to send messages to other peers on the same topic
	topic, err := pubSub.Join(TopicName)
	if err != nil {
		log.Fatal(err)
	}

	// subscribe to the topic to receive/ listen messages sent by other peers
	subscribe, err := topic.Subscribe()
	if err != nil {
		log.Fatal(err)
	}

	return &VerifierNode{
		host:  newHost,
		ps:    pubSub,
		topic: topic,
		sub:   subscribe,
	}
}

func (vn *VerifierNode) PrintHostInfo() {
	fmt.Println("Node ID:", vn.host.ID())
	for _, addr := range vn.host.Addrs() {
		fmt.Println("Listening on:", addr.Encapsulate(ma.StringCast("/p2p/"+vn.host.ID().String())))
	}
}

func (vn *VerifierNode) ListenForMessages(ctx context.Context) {
	go func() { // go routine, to make it able to run concurrently with the rest of the program
		for {
			msg, err := vn.sub.Next(ctx)
			if msg.ReceivedFrom != vn.host.ID(){ // ignore messages coming from itself
				if err != nil {
					log.Println("Error reading from subscription:", err)
					continue
				}
				fmt.Printf("Received: %s\n", string(msg.Data))
			}	
		}
	}()
}

func (vn *VerifierNode) SendMessage(ctx context.Context, message string) {
	err := vn.topic.Publish(ctx, []byte(message))
	if err != nil {
		log.Println("Error publishing message:", err)
	}
}

func ChainSpecToJson(target *substrate.ChainSpecRes) (string, error) {
    jsonBytes, err := json.MarshalIndent(target, "", "  ")
    if err != nil {
        fmt.Println("Error marshaling JSON:", err)
        return "", err // return empty string if error
    }

    return string(jsonBytes), nil
}

func (vn *VerifierNode) ConnectBootNode(target_json string, boot_index int){
	var spec substrate.ChainSpecRes
	err := json.Unmarshal([]byte(target_json), &spec)
	if err != nil {
		fmt.Println("Failed to unmarshal target JSON:", err)
		panic(err)
	}

	// Initialize connector with no cache (for clean testing)
	connector := substrate.NewSubstrateConnector(false)

	// Call the function
	api, err := connector.GetSubstrateApi(spec, boot_index)
	if err != nil {
		fmt.Println("Failed to get Substrate API:", err)
		panic(err)
	}

	// System info
	chainName, err := api.RPC.System.Chain()
	if err != nil {
		fmt.Printf("Failed to get chain name: %v\n", err)
		panic(err)
	}

	nodeName, err := api.RPC.System.Name()
	if err != nil {
		fmt.Printf("Failed to get node name: %v\n", err)
		panic(err)
	}

	nodeVersion, err := api.RPC.System.Version()
	if err != nil {
		fmt.Printf("Failed to get node version: %v\n", err)
		panic(err)
	}

	nodePeers, err := api.RPC.System.Peers()
	if err != nil {
		fmt.Printf("Failed to get peers: %v\n", err)
		panic(err)
	}

	blockHash, err := api.RPC.Chain.GetBlockHashLatest()
	if err != nil {
		fmt.Printf("Failed to get latest block hash: %v\n", err)
		panic(err)
	}

	fmt.Println("Successfully connected to Substrate API")
	fmt.Printf("Chain: %s\n", chainName)
	fmt.Printf("Node: %s\n", nodeName)
	fmt.Printf("Version: %s\n", nodeVersion)
	fmt.Printf("Latest block hash: %v\n", blockHash)

	fmt.Printf("Connected peers: %d\n", len(nodePeers))
	for i, peer := range nodePeers {
		fmt.Printf("Peer %d ID: %s, Role: %s\n", i+1, peer.PeerID, peer.Roles)
	}
}