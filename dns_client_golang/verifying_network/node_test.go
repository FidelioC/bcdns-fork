package verifying_network

import (
	"context"
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