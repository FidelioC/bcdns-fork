package verifying_network

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
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


func TestListenForMessages_validJSON(t *testing.T) {
	ctx := context.Background()
	topic := "test_topic_" + time.Now().Format("150405") // .Format() to format HH:mm:ss
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

func TestListenForMessages_InvalidJSON(t *testing.T) {
	ctx := context.Background()
	topic := "test_topic_" + time.Now().Format("150405")
	vn, _ := NewVerifierNode(ctx, "", 1, topic)
	defer vn.host.Close()

	errChan := make(chan error, 1)
	vn.OnError = func(err error) {
		errChan <- err
	}
	vn.ListenForMessages(ctx)

	// Send invalid JSON payload
	invalidPayload := []byte("{invalid-json")

	if err := vn.topic.Publish(ctx, invalidPayload); err != nil {
		t.Fatalf("Failed to publish invalid message: %v", err)
	}

	select {
	case err := <-errChan:
		if err == nil || !strings.Contains(err.Error(), "invalid JSON") {
			t.Errorf("Expected JSON error, got: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Timed out waiting for OnError for invalid JSON")
	}
}

func TestListenForMessages_ConsensusTrigger(t *testing.T) {
	ctx := context.Background()
	topic := "test_topic_" + time.Now().Format("150405")
	vn, _ := NewVerifierNode(ctx, "", 2, topic) // totalNodes = 2
	defer vn.host.Close()

	counter := 0
	vn.OnResult = func(res VerificationResult) {
		counter++
	}

	vn.ListenForMessages(ctx)

	// Send 2 messages with different NodeIDs
	for i := 0; i < 2; i++ {
		mockResult := VerificationResult{
			NodeID:    fmt.Sprintf("node-%d", i),
			ChainName: "TestChain",
			IsSuccess: true,
		}
		payload, _ := json.Marshal(mockResult)
		_ = vn.topic.Publish(ctx, payload)
	}

	time.Sleep(1 * time.Second)

	if len(vn.receivedResults) != 2 {
		t.Errorf("Expected 2 results, got %d", len(vn.receivedResults))
	}
}