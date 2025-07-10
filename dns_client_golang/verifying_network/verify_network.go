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

/* mostly used for testing and running a single node */
func RunSingleVerifier(bootstrap string){
	ctx := context.Background()
	vn := NewVerifierNode(ctx, bootstrap, 1)
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
	bootstrapNode := NewVerifierNode(ctx, "", n)
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
		node := NewVerifierNode(ctx, bootstrapAddr, n)
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

func ConnectBootNodes(nodes []*VerifierNode, json_target string, numNodes int) *VerificationResult{
	resultsChan := make(chan *VerificationResult, len(nodes))
	for i := range nodes {
		go func(i int) {
			// each of the result that's being returned by the node is a result of the consensus with the other peers
			res := nodes[i].GetBootNodeResult(context.Background(), json_target, 0, 30*time.Second)
			resultsChan <- res
		}(i)
	}
	
	var finalResult *VerificationResult
	for i := 0; i < numNodes; i++ {
		res := <-resultsChan
		fmt.Printf("Result %d:\n%+v\n", i, res)
		if res != nil && finalResult == nil {
			finalResult = res
		}
	}
	return finalResult
}

func GenerateResultJson(finalResult *VerificationResult, target *substrate.ChainSpecRes){
	if finalResult == nil {
		fmt.Println("No valid verification result received.")
	} else {
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
}

func VerifySpec(target *substrate.ChainSpecRes, numNodes int){
	json_target, err := ChainSpecToJson(target)

	if err != nil{
		panic(err)
	}

	// 1) create verifying network
	nodes := CreateVerifierNetwork(numNodes)

	// 2) connect to bootnodes and do consensus
	finalResult := ConnectBootNodes(nodes, json_target, numNodes)

	// 3) client collect results and generate to json file
	GenerateResultJson(finalResult, target)
}