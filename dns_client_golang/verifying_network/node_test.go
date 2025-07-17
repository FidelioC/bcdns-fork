package verifying_network

import (
	"context"
	"encoding/json"
	"testing"
	"time"
)

func TestNewVerifierNode(t *testing.T) {
	ctx := context.Background()
	topic := "test_topic_" + time.Now().Format("150405") // unique topic name
	totalNodes := 1

	node, err := NewVerifierNode(ctx, "", totalNodes, topic)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if node == nil {
		t.Fatal("Expected non-nil VerifierNode")
	}
	if node.totalNodes != totalNodes {
		t.Errorf("Expected totalNodes = %d, got %d", totalNodes, node.totalNodes)
	}
}


func TestListenForMessages(t *testing.T) {
	ctx := context.Background()
	topic := "test_topic_" + time.Now().Format("150405") // unique topic name
	// Create a test verifier node
	vn, _ := NewVerifierNode(ctx, "", 1, topic)
	defer vn.host.Close() // clean up libp2p host

	// Set up result capture channel
	resultChan := make(chan VerificationResult, 1)

	// Inject test handler
	vn.OnResult = func(res VerificationResult) {
		resultChan <- res
	}

	// Start listening
	vn.ListenForMessages(ctx)

	// Craft a dummy result
	mockResult := VerificationResult{
		NodeID:    vn.host.ID().String(),
		ChainName: "TestChain",
		NodeName:  "Node123",
		Version:   "1.0",
		BlockHash: "0xabc123",
		BootIndex: 0,
		IsSuccess: true,
	}
	payload, _ := json.Marshal(mockResult)

	// Publish the message to the same topic
	err := vn.topic.Publish(ctx, payload)
	if err != nil {
		t.Fatalf("Failed to publish test message: %v", err)
	}

	// Wait for OnResult or timeout
	select {
	case result := <-resultChan:
		if result.ChainName != "TestChain" {
			t.Errorf("Expected ChainName 'TestChain', got '%s'", result.ChainName)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Timed out waiting for OnResult")
	}
}