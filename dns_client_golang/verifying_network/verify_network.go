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

	"github.com/FidelioC/circom_zksnark/scripts"
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

func CreateVerifierNetwork(n int, mock substrate.APIInterface) []*VerifierNode {
	ctx := context.Background()
	nodes := make([]*VerifierNode, 0, n)

	bootstrapNode, _ := NewVerifierNode(ctx, "", n, TopicName)
	if mock != nil {
		bootstrapNode.MockAPI = mock
	}
	nodes = append(nodes, bootstrapNode)
	time.Sleep(1 * time.Second)

	var bootstrapAddr string
	for _, addr := range bootstrapNode.host.Addrs() {
		bootstrapAddr = addr.Encapsulate(ma.StringCast("/p2p/" + bootstrapNode.host.ID().String())).String()
		break
	}

	for i := 1; i < n; i++ {
		node, _ := NewVerifierNode(ctx, bootstrapAddr, n, TopicName)
		if mock != nil {
			node.MockAPI = mock
		}
		nodes = append(nodes, node)
		time.Sleep(500 * time.Millisecond)
	}

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
			resultsChan <- res
			errorsChan <- err
		}(i)
	}

	var finalResult *VerificationResult
	var hasError error

	for i := 0; i < numNodes; i++ {
		res := <-resultsChan
		err := <-errorsChan

		if err != nil {
			hasError = err // capture any error
		}
		if res != nil && finalResult == nil {
			finalResult = res
		}
	}

	// Return result even if some peers failed
	return finalResult, hasError
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

func VerifySpec(json_target string, numNodes int, mock_nodes []*VerifierNode) error {
	var nodes []*VerifierNode
	// 1) create verifying network
	if mock_nodes == nil{
		nodes = CreateVerifierNetwork(numNodes, nil)
	} else{
		nodes = mock_nodes
	}

	// 2) connect to bootnodes and do consensus
	finalResult, err := ConnectBootNodes(nodes, json_target, numNodes)
	if err != nil {
		return fmt.Errorf("failed during consensus phase: %w", err)
	}

	if finalResult == nil {
		return fmt.Errorf("no consensus reached among verifier nodes")
	}

	// 3) client collect results and generate to json file
	GenerateResultJson(finalResult, json_target)

	return nil
}

func ZkSnark_prove(domain_name string, input_json_path string){
	scripts.InitProve(domain_name)
	prove_folder := scripts.GenerateProve(domain_name, input_json_path)
	scripts.VerifyProve("--tld", prove_folder)
	scripts.Cleanup()
}

