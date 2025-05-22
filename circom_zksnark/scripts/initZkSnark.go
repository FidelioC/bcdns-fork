package main

// Usage: go run initZkSnark.go example.com

import (
	"fmt"
	"go_modules/go_modules"
	"os"
)

const (
	INIT_FOLDER_NAME = "../circom_inputs_init"
	TLD_FILE_NAME  = "../circom_inputs_init/tldSpecInput.json"
	TARGET_FILE_NAME  = "../circom_inputs_init/targetSpecInput.json"
	FORMAT_RESULT_INPUT = "../config/validChainSpec.json"
	FORMAT_RESULT_OUTPUT = "../circom_inputs_init/validSpecResult.json"
	TLD_CIRCOM_INPUT = "../circom_inputs_init/tldCircomInput.json"
	TARGET_CIRCOM_INPUT = "../circom_inputs_init/targetCircomInput.json"
	PROVE_FOLDER_NAME = "../circom_proves"
	PROVE_FORMAT_OUTPUT = "../circom_proves/proveSpecResult.json"
	PROVE_NAME_ID_OUTPUT = "../circom_proves/proveNameIdInput.json"
	PROVE_CIRCOM_INPUT = "../circom_proves/proveCircomInput.json"
)

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

		go_modules.CreateFolder(INIT_FOLDER_NAME)

		// create tld and target spec input json file based off the domain name
		go_modules.InitSpecInputPy(domain, TLD_FILE_NAME, TARGET_FILE_NAME)

		// create format binaries json
		go_modules.CheckFormatSpecPy(FORMAT_RESULT_INPUT, FORMAT_RESULT_OUTPUT)

		// combine both files spec (main & id) + (format binaries) to create circom json input
		go_modules.CombineJSONFile(TLD_FILE_NAME, FORMAT_RESULT_OUTPUT, TLD_CIRCOM_INPUT)
		go_modules.CombineJSONFile(TARGET_FILE_NAME, FORMAT_RESULT_OUTPUT, TARGET_CIRCOM_INPUT)

		// create circom prove
		go_modules.CallBashFiles("./runCircom.sh", "../tldCircomInput", "../.././circom_inputs_init/tldCircomInput.json")
		go_modules.CallBashFiles("./runCircom.sh", "../targetCircomInput", "../.././circom_inputs_init/targetCircomInput.json")

		fmt.Println("Finished init zk-snark circom")

	} else if mode == "--cleanup" {
		go_modules.CallBashFiles("./cleanCircomFiles.sh", "tldCircomInput")
		go_modules.CallBashFiles("./cleanCircomFiles.sh", "targetCircomInput")

		fmt.Println("Finished cleaning up")
	} else if mode == "--node-prove" {
		if len(os.Args) < 3 {
			fmt.Println("Usage: go run main.go --node-prove <input_file_to_be_proved>")
			return
		}
		go_modules.CreateFolder(PROVE_FOLDER_NAME)

		prove_file_path := os.Args[2]

		// create format binaries json
		go_modules.CheckFormatSpecPy(prove_file_path, PROVE_FORMAT_OUTPUT)

		// get the name and id field, convert to int
		result, _ := go_modules.FieldToInt(prove_file_path)
		go_modules.WriteJSONToFile(result, PROVE_NAME_ID_OUTPUT)

		// combine both files spec (main & id) + (format binaries) to create circom json input
		go_modules.CombineJSONFile(PROVE_NAME_ID_OUTPUT, PROVE_FORMAT_OUTPUT, PROVE_CIRCOM_INPUT)
		
		go_modules.CallBashFiles("./runCircom.sh", "../proveCircomFiles", "../.././circom_proves/proveCircomInput.json")

		fmt.Println("Finished creating proves zk-snark circom")
	} else {
		fmt.Println("Usage: go run initZkSnark.go --init <domain> OR go run initZkSnark.go --cleanup")
	}
}	
