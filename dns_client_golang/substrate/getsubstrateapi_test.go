package substrate

import (
	"encoding/json"
	"os"
	"testing"
)


func TestGetSubstrateApi(t *testing.T) {
	// Load spec from JSON
	data, err := os.ReadFile("../../polkadot-sdk-solochain-template/all_specs/com_tldSpec.json")
	if err != nil {
		t.Fatal("Failed to read spec JSON:", err)
	}

	var spec ChainSpecRes
	err = json.Unmarshal(data, &spec)
	if err != nil {
		t.Fatal("Failed to unmarshal spec JSON:", err)
	}

	// Initialize connector with no cache (for clean testing)
	connector := NewSubstrateConnector(false)

	// Call the function
	api, err := connector.getSubstrateApi(spec, 0)
	if err != nil {
		t.Fatal("Failed to get Substrate API:", err)
	}

	t.Log("Successfully connected to Substrate API:", api)
}
