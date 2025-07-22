package verifying_network

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/khalidzahra/dns_client/substrate"
	ma "github.com/multiformats/go-multiaddr"
)

// topic name should always be the same for all nodes in the verifying network
const TopicName = "verifying-network" 

/* mostly used for testing and running a single node */
func RunSingleVerifier(bootstrap string){
	ctx := context.Background()
	vn, _ := NewVerifierNode(ctx, bootstrap, 1, TopicName)
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
	bootstrapNode, _ := NewVerifierNode(ctx, "", n, TopicName)
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
		node, _ := NewVerifierNode(ctx, bootstrapAddr, n, TopicName)
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

func ConnectBootNodes(nodes []*VerifierNode, jsonTarget string, numNodes int) (*VerificationResult, error) {
	resultsChan := make(chan *VerificationResult, len(nodes))
	errorsChan := make(chan error, len(nodes))

	for i := range nodes {
		go func(i int) {
			res, err := nodes[i].GetBootNodeResult(context.Background(), jsonTarget, 0, 30*time.Second)
			if err != nil {
				errorsChan <- err
				resultsChan <- nil
			} else {
				errorsChan <- nil
				resultsChan <- res
			}
		}(i)
	}

	var finalResult *VerificationResult
	for i := 0; i < numNodes; i++ {
		res := <-resultsChan
		err := <-errorsChan

		if err != nil {
			// If it's a JSON issue or connection issue, propagate it back
			return nil, err
		}
		if res != nil && finalResult == nil {
			finalResult = res
		}
	}

	return finalResult, nil
}


func GenerateResultJson(finalResult *VerificationResult, json_target string){
	if finalResult == nil {
		fmt.Println("No valid verification result received.")
		return
	} 
	
	// Parse the target JSON
	var target substrate.ChainSpecRes
	if err := json.Unmarshal([]byte(json_target), &target); err != nil {
		log.Fatalf("Failed to parse target JSON: %v", err)
	}


	// Define minimal combined structure
	combined := struct {
		ChainName string   `json:"name"`
		ID        string   `json:"id"`
		BootNodes []string `json:"bootNodes"`
	}{
		ChainName: finalResult.ChainName,
		ID:        target.Id,
		BootNodes: target.BootNodes,
	}

	// Write to file
	file, err := os.Create("combined_result.json")
	if err != nil {
		log.Fatalf("Failed to create output file: %v", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(combined); err != nil {
		log.Fatalf("Failed to encode combined result: %v", err)
	}

	fmt.Println("Wrote combined_result.json successfully.")
	
}

func VerifySpec(json_target string, numNodes int) error {
	// 1) create verifying network
	nodes := CreateVerifierNetwork(numNodes)

	// 2) connect to bootnodes and do consensus
	finalResult, err := ConnectBootNodes(nodes, json_target, numNodes)
	if err != nil {
		return fmt.Errorf("failed during consensus phase: %w", err)
	}
	// 3) client collect results and generate to json file
	GenerateResultJson(finalResult, json_target)

	return nil
}
