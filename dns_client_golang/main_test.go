package main

import (
	"testing"
	"time"

	"github.com/khalidzahra/dns_client/substrate"
	"github.com/khalidzahra/dns_client/substrate/mocks"
)

// TestFetchSingleSpec tests the fetchSingleSpec function
func TestFetchSingleSpec(t *testing.T) {
	// Create test cases
	testCases := []struct {
		name     string
		domain   string
		idx      int
		evalFlag bool
	}{
		{
			name:     "Simple domain without eval",
			domain:   "example.com",
			idx:      5,
			evalFlag: false,
		},
		{
			name:     "Simple domain with eval",
			domain:   "example.com",
			idx:      10,
			evalFlag: true,
		},
		{
			name:     "Subdomain without eval",
			domain:   "sub.example.com",
			idx:      15,
			evalFlag: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a mock connector with controlled behavior
			mockConnector := &mocks.MockSubstrateConnector{
				ResolveDomainFunc: func(domain string, eval bool) (*substrate.ChainSpecRes, error) {
					// Add a small delay to simulate network latency
					time.Sleep(50 * time.Millisecond)

					// Verify the domain and eval flag are passed correctly
					if domain != tc.domain {
						t.Errorf("Expected domain %s, got %s", tc.domain, domain)
					}
					if eval != tc.evalFlag {
						t.Errorf("Expected eval flag %v, got %v", tc.evalFlag, eval)
					}

					// Return a mock chain spec
					return &substrate.ChainSpecRes{
						Id:        "test-chain-" + domain,
						BootNodes: []string{"/ip4/127.0.0.1/tcp/9944/ws"},
					}, nil
				},
			}

			// Call the function under test
			idx, duration, _ := fetchSingleSpec(tc.domain, tc.idx, mockConnector, tc.evalFlag, false)

			// Verify results
			if idx != tc.idx {
				t.Errorf("Expected index %d, got %d", tc.idx, idx)
			}

			// Duration should be greater than 0
			if duration <= 0 {
				t.Errorf("Expected positive duration, got %d", duration)
			}

			// The duration should be at least the sleep time
			if duration < 50 {
				t.Errorf("Expected duration to be at least 50ms, got %d ms", duration)
			}
		})
	}

	// Test error handling
	t.Run("Error handling", func(t *testing.T) {
		// Create a mock connector that returns an error
		mockConnector := &mocks.MockSubstrateConnector{
			ResolveDomainFunc: func(domain string, eval bool) (*substrate.ChainSpecRes, error) {
				return nil, &substrate.TestError{Message: "test error"}
			},
		}

		// Set up a recovery function to catch the panic
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("Expected panic but none occurred")
			}
		}()

		// This should panic
		fetchSingleSpec("example.com", 1, mockConnector, false, false)
	})
}
