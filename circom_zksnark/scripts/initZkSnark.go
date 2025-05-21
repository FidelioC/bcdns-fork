package main

// Usage: go run initZkSnark.go example.com

import (
	"fmt"
	"os"
	"os/exec"
)

const (
	FOLDER_NAME = "../circom_inputs"
	TLD_FILE_NAME  = "../circom_inputs/tldSpecInput.json"
	TARGET_FILE_NAME  = "../circom_inputs/targetSpecInput.json"
	FORMAT_RESULT_INPUT = "../config/validChainSpec.json"
	FORMAT_RESULT_OUTPUT = "../circom_inputs/validSpecResult.json"
	TLD_CIRCOM_INPUT = "../circom_inputs/tldCircomInput.json"
	TARGET_CIRCOM_INPUT = "../circom_inputs/targetCircomInput.json"
)

func printErrMessage(output []byte, err error){
	fmt.Println("Error:", err)
	fmt.Println("Output:\n", string(output))
}

func initSpecInputPy(domain string, tldOutput string, targetOutput string){
	cmd := exec.Command("python3", "initSpecInput.py", "--domain", domain, "--tld-output", tldOutput, "--target-output", targetOutput)

	output, err := cmd.CombinedOutput()
	if err != nil{
		printErrMessage(output, err)
		return
	}

	fmt.Println(string(output))
}

func checkFormatSpecPy(formatInput string, formatResultOutput string){
	cmd := exec.Command("python3", "checkSpecFormat.py", "--input-file", formatInput, "--output-file", formatResultOutput)

	output, err := cmd.CombinedOutput()
	if err != nil{
		printErrMessage(output, err)
		return
	}

	fmt.Println(string(output))
}

func combineJSONFile(file1 string, file2 string, combineFileOutput string){
	cmd := exec.Command("python3", "combineJson.py", "--file1", file1, "--file2", file2, "--output", combineFileOutput)

	output, err := cmd.CombinedOutput()

	if err != nil{
		printErrMessage(output, err)
		return
	}

	fmt.Println(string(output))
}

func callBashFiles(scriptPath string, args ...string){
	cmd := exec.Command("bash", append([]string{scriptPath}, args...)...)

	// connect scripts stdin, stdout and stderr, to Go's terminal
	// in order for it to be able to receive user's input
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		fmt.Printf("Error executing script %s: %v\n", scriptPath, err)
		return
	}
	return
}

func main(){
	if len (os.Args) < 2{
		fmt.Println("Usage: go run initZkSnark.go --init <domain> OR go run initZkSnark.go --cleanup")
		return
	}
	
	mode := os.Args[1]

	if mode == "--init"{
		if len(os.Args) < 3 {
			fmt.Println("Usage: go run main.go --init <domain>")
			return
		}
		// domain name
		domain := os.Args[2]

		err := os.Mkdir(FOLDER_NAME, 0755) // 0755 = rwxr-xr-x
		if err != nil {
			fmt.Printf("Error creating folder %s: %v\n", FOLDER_NAME, err)
			return
		}
		fmt.Printf("Folder: %s created successfully\n\n", FOLDER_NAME)

		// create tld and target spec input json file based off the domain name
		initSpecInputPy(domain, TLD_FILE_NAME, TARGET_FILE_NAME)

		// expected valid format json
		checkFormatSpecPy(FORMAT_RESULT_INPUT, FORMAT_RESULT_OUTPUT)

		// combine both files above to create circom json input
		combineJSONFile(TLD_FILE_NAME, FORMAT_RESULT_OUTPUT, TLD_CIRCOM_INPUT)
		combineJSONFile(TARGET_FILE_NAME, FORMAT_RESULT_OUTPUT, TARGET_CIRCOM_INPUT)

		// create circom prove
		callBashFiles("./runCircom.sh", "tldCircomInput")
		callBashFiles("./runCircom.sh", "targetCircomInput")

		fmt.Println("Finished init zk-snark circom")

	} else if mode == "--cleanup" {
		callBashFiles("./cleanCircomFiles.sh", "tldCircomInput")
		callBashFiles("./cleanCircomFiles.sh", "targetCircomInput")

		fmt.Println("Finished cleaning up")
	} else {
		fmt.Println("Usage: go run initZkSnark.go --init <domain> OR go run initZkSnark.go --cleanup")
	}
		
	
	
}	
