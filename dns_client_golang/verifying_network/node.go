// ================================================================
// File: node.go
// Description: Defines a VerifierNode which uses libp2p and gossip
//   pubsub to verify boot nodes in a decentralized Substrate network.
//   It connects to a boot node, fetches metadata (chain name, node
//   name, version, latest block hash), broadcasts the results, and
//   performs consensus across multiple peers.
// Author: Fidelio Ciandy
// =================================================================

package verifying_network

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/khalidzahra/dns_client/substrate"
	libp2p "github.com/libp2p/go-libp2p"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	host "github.com/libp2p/go-libp2p/core/host"
	peer "github.com/libp2p/go-libp2p/core/peer"
	peerstore "github.com/libp2p/go-libp2p/core/peerstore"
	ma "github.com/multiformats/go-multiaddr" //https://github.com/multiformats/multiaddr
)

// VerificationResult holds the result of connecting to a boot node and
// fetching its metadata for comparison/consensus
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

// VerifierNode represents a libp2p peer node capable of verifying
// a boot node, broadcasting its findings, and participating in consensus
type VerifierNode struct {
	host  host.Host // the libp2p host for the node
	ps    *pubsub.PubSub // GossipSub instance
	topic *pubsub.Topic // pubsub topic
	sub   *pubsub.Subscription // subscription to the topic
	receivedResults map[string]VerificationResult // results received from peers
	consensusAchieved bool // flag for whether consensus was reached
	consensusResult *VerificationResult // the result agreed upon
	totalNodes int  // expected number of participating nodes
	
	// MockAPI for testing
	MockAPI substrate.APIInterface

	// Function fields for easier testing
	OnResult func(result VerificationResult)
	OnError  func(err error)
}

// ============================================================================
// VerifierNode Initialization and Setup
// ============================================================================

// NewVerifierNode creates and initializes a new libp2p peer node, optionally
// connecting it to a bootstrap peer via multiaddress string
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

// ============================================================================
// Core Messaging + Consensus
// ============================================================================

// ListenForMessages continuously listens for messages on the pubsub topic,
// deserializes them, logs and stores them, and optionally triggers consensus
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

// CheckConsensus analyzes all received results and sets the consensus result
// if a simple majority agrees on the same chain metadata
func (vn *VerifierNode) CheckConsensus() {
	type key struct {
		ChainName string
		Version   string
		BlockHash string
	}

	// Count matching results
	counts := make(map[key]int)
	keyToResult := make(map[key]VerificationResult) // store one representative result for each key

	for _, res := range vn.receivedResults {
		if res.IsSuccess {
			k := key{res.ChainName, res.Version, res.BlockHash}
			counts[k]++
			// Store one of the successful results per key (any one is enough)
			if _, exists := keyToResult[k]; !exists {
				keyToResult[k] = res
			}
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

	// Accept consensus if it meets quorum (simple majority)
	if maxCount > vn.totalNodes/2 {
		vn.consensusAchieved = true
		consensusRes := keyToResult[consensusKey]
		vn.consensusResult = &consensusRes
	}
}

// SendMessage broadcasts a string message to all peers via pubsub
func (vn *VerifierNode) SendMessage(ctx context.Context, message string) {
	err := vn.topic.Publish(ctx, []byte(message))
	if err != nil {
		log.Println("Error publishing message:", err)
	}
}

// ============================================================================
// Boot Node Verification
// ============================================================================
// ConnectBootNode attempts to connect to a Substrate boot node using the
// provided JSON chain spec and retrieves its metadata
func (vn *VerifierNode) ConnectBootNode(ctx context.Context, targetJSON string, bootIndex int) error {
	var spec substrate.ChainSpecRes
	err := json.Unmarshal([]byte(targetJSON), &spec)
	if err != nil {
		return fmt.Errorf("failed to unmarshal target JSON: %w", err)
	}

	connector := substrate.NewSubstrateConnector(false)

	var api substrate.APIInterface
	if vn.MockAPI != nil {
		// Use mock in tests
		api = vn.MockAPI
	} else {
		realAPI, err := connector.GetSubstrateApi(spec, bootIndex)
		if err != nil {
			return fmt.Errorf("API connection failed: %w", err)
		}
		api = &substrate.APIWrapper{API: realAPI}
	}

	result := VerificationResult{
		NodeID:    vn.host.ID().String(),
		BootIndex: bootIndex,
		IsSuccess: true,
	}

	// Try to get metadata
	result_chain, err1 := api.Chain()
	result_nodename, err2 := api.Name()
	result_version, err3 := api.Version()
	blockHash, err4 := api.GetBlockHashLatest()

	// Handle metadata errors
	if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
		result.IsSuccess = false
		result.ErrorMsg = fmt.Sprintf("Metadata fetch errors: Chain=%v, Name=%v, Version=%v, BlockHash=%v",
			err1, err2, err3, err4)
	} else {
		result.ChainName = convertTextToString(result_chain)
		result.NodeName = convertTextToString(result_nodename)
		result.Version = convertTextToString(result_version)
		result.BlockHash = blockHash.Hex()
	}

	// Marshal and broadcast
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to marshal verification result: %w", err)
	}

	vn.SendMessage(ctx, string(resultJSON))
	return nil
}

// GetBootNodeResult triggers a connection to the boot node and waits for a
// consensus result or timeout before returning
func (vn *VerifierNode) GetBootNodeResult(ctx context.Context, targetJSON string, bootIndex int, timeout time.Duration) (*VerificationResult, error) {
	// 1) connect to boot node, this function will also broadcast the result to other peers
	err := vn.ConnectBootNode(ctx, targetJSON, bootIndex)
	if err != nil {
		return nil, fmt.Errorf("connect bootnode failed: %w", err)
	}

	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()
	timeoutChan := time.After(timeout)

	// 2) will try and timeout if consensus hasn't been reached for a certain time
	for {
		select {
		case <-ticker.C:
			if vn.consensusResult != nil {
				return vn.consensusResult, nil
			}
		case <-timeoutChan:
			fmt.Printf("Timeout [Node %s] : No consensus achieved\n", vn.host.ID().String())
			return nil, nil
		}
	}
}



// ============================================================================
// Utility Functions
// ============================================================================

func (vn *VerifierNode) PrintHostInfo() {
	fmt.Println("Node ID:", vn.host.ID())
	for _, addr := range vn.host.Addrs() {
		fmt.Println("Listening on:", addr.Encapsulate(ma.StringCast("/p2p/"+vn.host.ID().String())))
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