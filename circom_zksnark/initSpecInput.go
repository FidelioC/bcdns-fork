package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type SpecInput struct{
	Name uint64 `json:"name"`
	ID uint64 `json:"id"`
}

func createSpecJSON(tldInput uint64) (SpecInput){
	SpecJSON := SpecInput{
		Name: tldInput,
		ID: tldInput,
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

func SplitDomain(domainInput string) (string, string, error){
	split := strings.Split(domainInput, ".")

	if len(split) != 2 {
		return "", "", fmt.Errorf("invalid website/ domain format: %s", domainInput)
	}

	return split[0], split[1], nil
}

func stringToNumber(stringInput string) uint64 {
	// base 256 encoding
	var result uint64 = 0;
	for i:=0; i < len(stringInput); i++{
		result = result*256 + uint64(stringInput[i])
	}

	return result
}

func CreateJSONFile(stringInput string, fileName string){
	var numberEncoding uint64 = stringToNumber(stringInput)
	SpecInput := createSpecJSON(numberEncoding)
	createSpecJSONFile(SpecInput, fileName)
}

func initSpecInput(domain string, tldFileName string, targetFileName string){
	// split domain
	domainName, tld, err := SplitDomain(domain)
	
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}

	fmt.Println("Input:", domain)
	fmt.Println("Domain name:", domainName)
	fmt.Println("tld:", tld)
	
	// create tld spec input json
	CreateJSONFile(tld, tldFileName)

	// create target spec input json
	CreateJSONFile(domainName, targetFileName)
}

func checkFormatSpecPy(formatInput string, formatResultOutput string){
	cmd := exec.Command("python3", "checkSpecFormat.py", "--input-file", formatInput, "--output-file", formatResultOutput)

	output, err := cmd.CombinedOutput()

	if err != nil{
		fmt.Println("Error:", err)
		fmt.Println("Python output:\n", string(output))
		return
	}

	fmt.Println(string(output))
}

func combineJSONFile(file1 string, file2 string, combineFileOutput string){
	cmd := exec.Command("python3", "combineJson.py", "--file1", file1, "--file2", file2, "--output", combineFileOutput)

	output, err := cmd.CombinedOutput()

	if err != nil{
		fmt.Println("Error:", err)
		fmt.Println("Python output:\n", string(output))
		return
	}

	fmt.Println(string(output))
}

func main(){
	if len (os.Args) < 2{
		fmt.Println("Usage: go run main.go <domain>")
		return
	}
	
	// domain name
	domain := os.Args[1]

	// create tld and target spec input json file based off the domain name
	tldFileName := "tldSpecInput.json" // output
	targetFileName := "targetSpecInput.json" // output
	initSpecInput(domain, tldFileName, targetFileName)

	// expected valid format json
	formatResultInput := "validChainSpec.json"
	formatResultOutput := "validSpecResult.json" // output
	checkFormatSpecPy(formatResultInput, formatResultOutput)

	// combine both files above to create circom json input
	tldCircomInput := "tldCircomInput.json" // output
	targetCircomInput := "targetCircomInput.json" // output
	combineJSONFile(tldFileName, formatResultOutput, tldCircomInput)
	combineJSONFile(targetFileName, formatResultOutput, targetCircomInput)
}	
