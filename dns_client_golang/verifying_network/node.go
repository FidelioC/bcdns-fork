package verifying_network

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

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

type VerifierNode struct {
	host  host.Host
	ps    *pubsub.PubSub
	topic *pubsub.Topic
	sub   *pubsub.Subscription
	receivedResults map[string]VerificationResult
	consensusAchieved bool
	consensusResult *VerificationResult
	totalNodes int

	// Function fields for easier testing
	OnResult func(result VerificationResult)
	OnError  func(err error)
}

func NewVerifierNode(ctx context.Context, bootstrap string, totalNodes int, nameTopic string) (*VerifierNode, error) {
	// create the libp2p node (host)
	newHost, err := libp2p.New() 
	if err != nil {
		return nil, fmt.Errorf("failed to create libp2p host: %w", err)
	}

	// if bootstrap address was provided, to talk to other existing peers
	if bootstrap != "" {
		// parse the multiaddr string, e.g., /ip6/2604:3d09:a98d:b100:4552:be06:a3ca:b295/udp/64042/webrtc-direct/certhash/uEiAAKlDpHtl0D3aOBbJEEYqArQLTuZ9zL-smFMJ17JGrag/p2p/12D3KooWAVaoXdP8wurmgFXizqKHV4NGZnHzJxpKdmyAvfS9tEW1
		maddr, err := ma.NewMultiaddr(bootstrap)
		if err != nil {
			return nil, fmt.Errorf("invalid bootstrap multiaddr: %w", err)
		}

		// convert the maddr to peer.Addrinfo, which separates the peer ID and addresses
		info, err := peer.AddrInfoFromP2pAddr(maddr)
		if err != nil {
			return nil, fmt.Errorf("failed to extract AddrInfo: %w", err)
		}
		
		// add the address to this node's internal peer storage
		newHost.Peerstore().AddAddrs(info.ID, info.Addrs, peerstore.PermanentAddrTTL)

		// connect to peer
		newHost.Connect(ctx, *info)
	}
	// initialize new GossipSub
	pubSub, err := pubsub.NewGossipSub(ctx, newHost)
	if err != nil {
		return nil, fmt.Errorf("failed to create pubsub: %w", err)
	}

	// join on a pubsub topic (constant), to send messages to other peers on the same topic
	topic, err := pubSub.Join(nameTopic)
	if err != nil {
		return nil, fmt.Errorf("failed to join pubsub topic: %w", err)
	}

	// subscribe to the topic to receive/ listen messages sent by other peers
	subscribe, err := topic.Subscribe()
	if err != nil {
		return nil, fmt.Errorf("failed to subscribe to pubsub topic: %w", err)
	}

	return &VerifierNode{
		host:  newHost,
		ps:    pubSub,
		topic: topic,
		sub:   subscribe,
		receivedResults: make(map[string]VerificationResult),
		totalNodes: totalNodes,
	}, nil
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
			// get the message
			msg, err := vn.sub.Next(ctx)
			if err != nil {
				vn.handleError(fmt.Errorf("error reading message: %w", err))
				continue
			}

			// convert msg to json
			var result VerificationResult
			err = json.Unmarshal(msg.Data, &result)
			
			if err != nil {
				vn.handleError(fmt.Errorf("invalid JSON: %w", err))
				continue
			}
			
			// store result and format to json
			vn.receivedResults[result.NodeID] = result

			// call optional test handler
			if vn.OnResult != nil {
				vn.OnResult(result)
			}

			jsonPretty, err := json.MarshalIndent(result, "", "  ")

			if err != nil {
				vn.handleError(fmt.Errorf("JSON formatting error: %w", err))
			} else {
				fmt.Printf("[Node %s] Received verification from %s:\n%s\n",
					vn.host.ID().String(), result.NodeID, string(jsonPretty))
				fmt.Printf("[Node %s] Current received result length: %v\n", vn.host.ID().String(), len(vn.receivedResults))
			}

			// Perform consensus check when enough results are collected
			if len(vn.receivedResults) >= vn.totalNodes {
				vn.CheckConsensus()
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

	var maxCount int
	var consensusKey key
	for k, c := range counts {
		if c > maxCount {
			maxCount = c
			consensusKey = k
		}
	}

	// Check if *this node* has a result and it agrees with the consensus
	selfRes, ok := vn.receivedResults[vn.host.ID().String()]
	if ok && selfRes.IsSuccess &&
		selfRes.ChainName == consensusKey.ChainName &&
		selfRes.Version == consensusKey.Version &&
		selfRes.BlockHash == consensusKey.BlockHash &&
		maxCount > vn.totalNodes/2 {
			vn.consensusAchieved = true
			vn.consensusResult = &selfRes
	}
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

func (vn *VerifierNode) GetBootNodeResult(ctx context.Context, targetJSON string, bootIndex int, timeout time.Duration) *VerificationResult {
	// 1) connect to boot node, this function will also broadcast the result to other peers
	vn.ConnectBootNode(ctx, targetJSON, bootIndex)

	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()
	timeoutChan := time.After(timeout)

	// 2) will try and timeout if consensus hasn't been reached for a certain time
	for {
		select {
		case <-ticker.C:
			if vn.consensusResult != nil {
				return vn.consensusResult
			}
		case <-timeoutChan:
			fmt.Printf("Timeout [Node %s] : No consensus achieved\n", vn.host.ID().String())
			return nil
		}
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

func convertTextToString(text types.Text) string{
	return string(text)
}

func (vn *VerifierNode) GetConsensusResult() (*VerificationResult){
	if vn.consensusAchieved {
		return vn.consensusResult
	} else {
		return nil
	}
}

func PrintVerifierNodeSummary(vn *VerifierNode) {
	fmt.Println("===== VerifierNode Summary =====")

	// Host ID
	fmt.Printf("Host ID: %s\n", vn.host.ID())

	// Host multiaddresses
	fmt.Println("Listening Addresses:")
	for _, addr := range vn.host.Addrs() {
		fmt.Printf(" - %s\n", addr.String())
	}

	// Topic (name not directly accessible, but can indicate presence)
	if vn.topic != nil {
		fmt.Println("PubSub Topic: [joined]")
	} else {
		fmt.Println("PubSub Topic: [not joined]")
	}

	// Subscription status
	if vn.sub != nil {
		fmt.Println("Subscribed: yes")
	} else {
		fmt.Println("Subscribed: no")
	}

	// Total nodes
	fmt.Printf("Total Nodes: %d\n", vn.totalNodes)

	// Received results count
	fmt.Printf("Received Results: %d\n", len(vn.receivedResults))

	// Consensus status
	if vn.consensusAchieved {
		fmt.Println("Consensus Achieved: True")
		if vn.consensusResult != nil {
			fmt.Printf("Consensus Result: %+v\n", *vn.consensusResult)
		}
	} else {
		fmt.Println("Consensus Achieved: False")
	}
}

func (vn *VerifierNode) handleError(err error) {
	if vn.OnError != nil {
		vn.OnError(err)
	}
}