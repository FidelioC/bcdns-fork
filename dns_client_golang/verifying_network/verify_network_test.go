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


