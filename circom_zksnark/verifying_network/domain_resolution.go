package verifying_network

// import (
// 	"fmt"
// 	"log"

// 	"github.com/khalidzahra/dns_client/substrate"
// )

// func DomainResolution(domainName string) {
// 	// Initialize the SubstrateConnector
// 	connector := substrate.NewSubstrateConnector(false)

// 	// Resolve a domain
// 	domain := domainName
// 	fmt.Printf("Resolving Domain: %s", domainName)
// 	chainSpec, err := connector.ResolveDomain(domain, false)
// 	if err != nil {
// 		log.Fatalf("Error resolving domain: %v", err)
// 	}

// 	fmt.Printf("Resolved Chain Spec: %+v\n", chainSpec)
// }