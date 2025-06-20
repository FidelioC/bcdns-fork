package verifying_network

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	ma "github.com/multiformats/go-multiaddr"
)

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