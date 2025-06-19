package verifying_network

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	// https://pkg.go.dev/github.com/libp2p/go-libp2p#section-readme

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

/* mostly used for testing and running a single node */
func RunSingleVerifier(bootstrap string){
	ctx := context.Background()
	vn := NewVerifierNode(ctx, bootstrap)
	vn.PrintHostInfo()
	vn.ListenForMessages(ctx)

	// Simple CLI loop to send messages to other nodes
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Type messages to send to other nodes:")
	for scanner.Scan() {
		text := scanner.Text()
		if strings.TrimSpace(text) == "" {
			continue
		}
		vn.SendMessage(ctx, fmt.Sprintf("%s: %s", vn.host.ID().ShortString(), text))
		time.Sleep(100 * time.Millisecond)
	}
}

func CreateVerifierNetwork(n int) []*VerifierNode {
	ctx := context.Background()
	nodes := make([]*VerifierNode, 0, n)

	// 1. Create the bootstrap node (no address needed)
	bootstrapNode := NewVerifierNode(ctx, "")
	nodes = append(nodes, bootstrapNode)
	fmt.Println("Bootstrap node started")
	bootstrapNode.PrintHostInfo()

	// Allow time for the bootstrap node to start
	time.Sleep(1 * time.Second)

	// 2. Grab the bootstrap node's full multiaddress
	var bootstrapAddr string
	for _, addr := range bootstrapNode.host.Addrs() {
		bootstrapAddr = addr.Encapsulate(ma.StringCast("/p2p/" + bootstrapNode.host.ID().String())).String()
		break
	}
	fmt.Println("Bootstrap address:", bootstrapAddr)

	// 3. Spin up remaining nodes, connecting to bootstrap
	for i := 1; i < n; i++ {
		node := NewVerifierNode(ctx, bootstrapAddr)
		nodes = append(nodes, node)
		node.PrintHostInfo()
		time.Sleep(500 * time.Millisecond)
		fmt.Printf("Node %d connected to bootstrap.\n", i)
	}

	// 4. Start listening for messages
	for _, node := range nodes {
		node.ListenForMessages(ctx)
	}

	return nodes
}
