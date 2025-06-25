package verifying_network

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	// https://pkg.go.dev/github.com/libp2p/go-libp2p#section-readme

	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/khalidzahra/dns_client/substrate"
	libp2p "github.com/libp2p/go-libp2p"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	host "github.com/libp2p/go-libp2p/core/host"
	peer "github.com/libp2p/go-libp2p/core/peer"
	peerstore "github.com/libp2p/go-libp2p/core/peerstore"
	ma "github.com/multiformats/go-multiaddr" //https://github.com/multiformats/multiaddr
)

type VerificationResult struct {
	NodeID      string `json:"nodeId"`
	ChainName   string `json:"chainName"`
	NodeName    string `json:"nodeName"`
	Version     string `json:"version"`
	BlockHash   string `json:"blockHash"`
	BootIndex   int    `json:"bootIndex"`
	IsSuccess   bool   `json:"isSuccess"`
	ErrorMsg    string `json:"errorMsg,omitempty"`
}

// topic name should always be the same for all nodes in the verifying network
const TopicName = "verifying-network" 

var (
	consensusAchieved bool = false
	consensusResult *VerificationResult
)
type VerifierNode struct {
	host  host.Host
	ps    *pubsub.PubSub
	topic *pubsub.Topic
	sub   *pubsub.Subscription

	receivedResults map[string]VerificationResult
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
		receivedResults: make(map[string]VerificationResult),
	}
}

func (vn *VerifierNode) PrintHostInfo() {
	fmt.Println("Node ID:", vn.host.ID())
	for _, addr := range vn.host.Addrs() {
		fmt.Println("Listening on:", addr.Encapsulate(ma.StringCast("/p2p/"+vn.host.ID().String())))
	}
}

func (vn *VerifierNode) ListenForMessages(ctx context.Context) {
	go func() {
		for {
			msg, err := vn.sub.Next(ctx)
			if err != nil {
				log.Println("Error reading message:", err)
				continue
			}
			if msg.ReceivedFrom == vn.host.ID() {
				continue
			}

			var result VerificationResult
			err = json.Unmarshal(msg.Data, &result)
			if err == nil {
				vn.receivedResults[result.NodeID] = result

				jsonPretty, err := json.MarshalIndent(result, "", "  ")
				if err == nil {
					fmt.Printf("[Node %s] Received verification from %s:\n%s\n",
						vn.host.ID().String(), result.NodeID, string(jsonPretty))
					fmt.Printf("[Node %s] Current received result length: %v\n", vn.host.ID().String(), len(vn.receivedResults))
				} else {
					fmt.Printf("Received result from %s, but failed to format JSON: %v\n", result.NodeID, err)
				}

				// Perform consensus check when enough results are collected
				if len(vn.receivedResults) >= 2 { // You can adjust this threshold
					vn.CheckConsensus()
				}
			} else {
				fmt.Printf("Received non-verification message: %s\n", string(msg.Data))
			}
		}
	}()
}

func (vn *VerifierNode) CheckConsensus() {
	type key struct {
		ChainName string
		Version   string
		BlockHash string
	}

	counts := make(map[key]int)
	for _, res := range vn.receivedResults {
		if res.IsSuccess {
			k := key{res.ChainName, res.Version, res.BlockHash}
			counts[k]++
		}
	}

	fmt.Printf("\n\n------ CONSENSUS REPORT NODE: %s ------\n", vn.host.ID().String())
	var maxCount int
	var consensusKey key
	for k, c := range counts {
		fmt.Printf("Config: %+v | Votes: %d\n", k, c)
		if c > maxCount {
			maxCount = c
			consensusKey = k
		}
	}

	if maxCount > len(vn.receivedResults)/2 {
		fmt.Printf("Consensus achieved: %+v\n\n", consensusKey)
		consensusAchieved = true
	} else {
		fmt.Println("No consensus reached.")
	}
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

func (vn *VerifierNode) ConnectBootNode(ctx context.Context, targetJSON string, bootIndex int) {
	var spec substrate.ChainSpecRes
	err := json.Unmarshal([]byte(targetJSON), &spec)
	if err != nil {
		log.Println("Failed to unmarshal target JSON:", err)
		return
	}

	connector := substrate.NewSubstrateConnector(false)

	api, err := connector.GetSubstrateApi(spec, bootIndex)
	result := VerificationResult{
		NodeID:    vn.host.ID().String(),
		BootIndex: bootIndex,
		IsSuccess: true, // assume success, set to false on any failure
	}

	if err != nil {
		result.ErrorMsg = fmt.Sprintf("API connection failed: %v", err)
		result.IsSuccess = false
	} else {
		// try to get metadata
		result_chain, err1 := api.RPC.System.Chain()
		result_nodename, err2 := api.RPC.System.Name()
		result_version, err3 := api.RPC.System.Version()
		blockHash, err4 := api.RPC.Chain.GetBlockHashLatest()

		// if any error occurs, mark failure
		if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
			result.IsSuccess = false
			result.ErrorMsg = fmt.Sprintf("Metadata fetch errors: Chain=%v, Name=%v, Version=%v, BlockHash=%v",
				err1, err2, err3, err4)
		} else {
			// populate result
			result.ChainName = convertTextToString(result_chain)
			result.NodeName = convertTextToString(result_nodename)
			result.Version = convertTextToString(result_version)
			result.BlockHash = blockHash.Hex()
		}
	}

	// Broadcast result
	resultJSON, err := json.Marshal(result)
	if err != nil {
		log.Println("Failed to marshal verification result:", err)
		return
	}

	vn.SendMessage(ctx, string(resultJSON))
}

func convertTextToString(text types.Text) string{
	return string(text)
}

func (vn *VerifierNode) GetConsensusResult() (*VerificationResult){
	if consensusAchieved {
		return consensusResult
	} else {
		return nil
	}
}
