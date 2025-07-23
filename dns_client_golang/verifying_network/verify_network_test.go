package verifying_network

import (
	"os"
	"testing"
)

func TestVerifySpec_HappyPath(t *testing.T) {

	jsonInput := `
	{
		"id": "example",
		"bootNodes": [
			"/ip4/172.20.0.2/tcp/9945/p2p/12D3KooWNL4mZo8y7oAes3VRRnbHy91TDLxnjrDsnMFZkPebB2Rh",
			"/ip4/172.20.0.3/tcp/9945/p2p/12D3KooWNL4mZo8y7oAes3VRRnbHy91TDLxnjrDsnMFZkPebB2Rh"
		]
	}`

	VerifySpec(jsonInput, 3) // numNodes = 3

	// Check if file was generated
	if _, err := os.Stat("combined_result.json"); os.IsNotExist(err) {
		t.Fatal("Expected output file combined_result.json was not created")
	}

	// Optional: clean up
	_ = os.Remove("combined_result.json")
}

func TestVerifySpec_InvalidJSON(t *testing.T) {
	// Malformed JSON (missing closing brace)
	invalidJSON := `{
		"id": "example",
		"bootNodes": [
			"/ip4/127.0.0.1/tcp/9945/p2p/QmSomeNode"
		]` // missing closing brace

	err := VerifySpec(invalidJSON, 3)
	if err == nil {
		t.Fatal("Expected error due to invalid JSON input, but got none")
	}
}

func TestVerifySpec_UnreachableBootNodes(t *testing.T) {
	// Point to fake bootnodes that are guaranteed to not resolve/connect
	unreachableJSON := `{
		"id": "example",
		"bootNodes": [
			"/ip4/192.0.2.123/tcp/9945/p2p/QmFakeNode1",
			"/ip4/192.0.2.124/tcp/9945/p2p/QmFakeNode2"
		]
	}`

	// Run the spec verification
	err := VerifySpec(unreachableJSON, 3)

	// We expect it to fail due to bootnode connection errors
	if err == nil {
		t.Fatal("Expected error due to unreachable boot nodes, but got none")
	}

	// Ensure no result file is generated
	if _, statErr := os.Stat("combined_result.json"); !os.IsNotExist(statErr) {
		t.Fatal("Output file should not exist when consensus fails")
	}
}




