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

	// Fetch and log chain info
	chainName, err := api.RPC.System.Chain()
	if err != nil {
		t.Fatalf("Failed to get chain name: %v", err)
	}

	nodeName, err := api.RPC.System.Name()
	if err != nil {
		t.Fatalf("Failed to get node name: %v", err)
	}

	nodeVersion, err := api.RPC.System.Version()
	if err != nil {
		t.Fatalf("Failed to get node version: %v", err)
	}

	blockHash, err := api.RPC.Chain.GetBlockHashLatest()
	if err != nil {
		t.Fatalf("Failed to get latest block hash: %v", err)
	}

	peerId, err := api.RPC.System.LocalPeerId()
	if err != nil {
		t.Fatalf("Failed to get peer ID: %v", err)
	}

	listenAddrs, err := api.RPC.System.LocalListenAddresses()
	if err != nil {
		t.Fatalf("Failed to get listen addresses: %v", err)
	}

	t.Logf("Successfully connected to Substrate API")
	t.Logf("Chain: %s", chainName)
	t.Logf("Node: %s", nodeName)
	t.Logf("Version: %s", nodeVersion)
	t.Logf("Latest block hash: %v", blockHash)
	t.Logf("Peer ID: %s", peerId)
	for i, addr := range listenAddrs {
		t.Logf("Listening address %d: %s", i+1, addr)
	}
}
