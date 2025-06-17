package utils

import (
	"fmt"
	"log"

	"github.com/khalidzahra/dns_client/substrate"
)

func DomainResolution(domainName string) {
	// Initialize the SubstrateConnector
	connector := substrate.NewSubstrateConnector(true)

	// Resolve a domain
	domain := domainName
	fmt.Println("Resolving Domain")
	chainSpec, err := connector.ResolveDomain(domain, true)
	if err != nil {
		log.Fatalf("Error resolving domain: %v", err)
	}

	fmt.Printf("Resolved Chain Spec: %+v\n", chainSpec)
}