package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type SpecInput struct{
	Name string `json:"name"`
	ID string `json:"id"`
	ChainType string `json:"chainType"`
	BootNodes *[]string `json:"bootNodes"`
	TelemetryEndpoints *[]string `json:"telemetryEndpoints"`
	ProtocolID *map[string]string `json:"protocolID"`
	Properties *map[string]string `json:"properties"`
	CodeSubstitutes map[string]string `json:"codeSubstitutes"`
}

func createSpecJSON(tldInput string) (SpecInput){
	SpecJSON := SpecInput{
		Name: tldInput,
		ID: tldInput,
		ChainType: "default",
		BootNodes: nil,
		TelemetryEndpoints: nil,
		ProtocolID: nil,
		Properties: nil,
		CodeSubstitutes: map[string]string{},
	}

	return SpecJSON
}

func createSpecJSONFile(jsonInput SpecInput, fileName string){
	// create JSON file
	file, err := os.Create(fileName)

	// file creation check
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer file.Close()

	//encode json and write to a file
	encoder := json.NewEncoder(file)
	err = encoder.Encode(jsonInput)

	// json encode check
	if err != nil {
		fmt.Println("Error encoding JSON:", err)
		return
	}

	fmt.Printf("%s created successfully\n", fileName)
}

func splitDomain(domainInput string) (string, string, error){
	split := strings.Split(domainInput, ".")

	if len(split) != 2 {
		return "", "", fmt.Errorf("invalid website/ domain format: %s", domainInput)
	}

	return split[0], split[1], nil
}

func main(){
	if len (os.Args) < 2{
		fmt.Println("Usage: go run main.go <domain>")
		return
	}

	domain := os.Args[1]

	// split domain
	domainName, tld, err := splitDomain(domain)
	
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}

	fmt.Println("Input:", domain)
	fmt.Println("Domain name:", domainName)
	fmt.Println("tld:", tld)
	
	// create tld spec input json
	tldSpecInput := createSpecJSON(tld)
	createSpecJSONFile(tldSpecInput, "tldSpecInput.json")

	// create target spec input json
	targetSpecInput := createSpecJSON(domainName)
	createSpecJSONFile(targetSpecInput, "targetSpecInput.json")
}