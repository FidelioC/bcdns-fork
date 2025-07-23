package verifying_network

import (
	"bytes"
	"os"
	"testing"

	"github.com/khalidzahra/dns_client/substrate/mocks"
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

	VerifySpec(jsonInput, 3, nil) // numNodes = 3

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

	err := VerifySpec(invalidJSON, 3, nil)
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
	err := VerifySpec(unreachableJSON, 3, nil)

	// We expect it to fail due to bootnode connection errors
	if err == nil {
		t.Fatal("Expected error due to unreachable boot nodes, but got none")
	}

	// Ensure no result file is generated
	if _, statErr := os.Stat("combined_result.json"); !os.IsNotExist(statErr) {
		t.Fatal("Output file should not exist when consensus fails")
	}
}

func TestVerifySpec_ConsensusOnWrongMetadata(t *testing.T) {
	numNodes := 3

	// 1. Create the verifier network with the mock bad API injected
	mock_nodes := CreateVerifierNetwork(numNodes, &mocks.MockBadAPI{})

	// 2. Prepare valid JSON input (boot node connects fine but metadata is fake)
	jsonInput := `{
		"id": "example",
		"bootNodes": [
			"/ip4/172.20.0.2/tcp/9945/p2p/12D3KooWNL4mZo8y7oAes3VRRnbHy91TDLxnjrDsnMFZkPebB2Rh",
			"/ip4/172.20.0.3/tcp/9945/p2p/12D3KooWNL4mZo8y7oAes3VRRnbHy91TDLxnjrDsnMFZkPebB2Rh"
		]
	}`

	// 3. Run the full VerifySpec (nodes use mock internally)
	err := VerifySpec(jsonInput, numNodes, mock_nodes)
	if err != nil {
		t.Fatalf("Unexpected error during VerifySpec: %v", err)
	}

	// 4. Check that the output file was created
	data, readErr := os.ReadFile("combined_result.json")
	if readErr != nil {
		t.Fatal("Expected output file combined_result.json was not created")
	}
	t.Logf("Generated Result JSON:\n%s", string(data))

	defer os.Remove("combined_result.json")

	// 5. Assert the wrong metadata was written
	if !bytes.Contains(data, []byte("WrongChain")) {
		t.Errorf("Expected output to contain 'WrongChain', got: %s", string(data))
	}
}






