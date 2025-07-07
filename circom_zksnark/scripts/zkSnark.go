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
	TLD_INPUT_FOLDER = "../tldCircomInput"
	TARGET_CIRCOM_INPUT = "../circom_inputs_init/targetCircomInput.json"
	TARGET_INPUT_FOLDER = "../targetCircomInput"
	PROVE_CIRCOM_INPUT_FOLDER = "../proveCircomFiles"
	PROVE_FORMAT_OUTPUT = "/proveSpecResult.json"
	PROVE_NAME_ID_OUTPUT = "/proveNameIdOutput.json"
	PROVE_CIRCOM_INPUT = "/proveCircomInput.json"
	VERIFICATION_KEY_PATH = "/circomCheck_js/verification_key.json"
	PUBLIC_JSON_PATH = "/circomCheck_js/public.json"
	PROOF_JSON_PATH = "/circomCheck_js/proof.json"
)

func main(){
	if len (os.Args) < 2{
		fmt.Println(`Usages: 
			- go run zkSnark.go --init <domain> 
			- go run zkSnark.go --node-generate-prove <input_json_file_to_be_proved>
			- go run zkSnark.go --verify-prove <--tld or --target> <prove_folder_to_be_verified>
			- go run zkSnark.go --cleanup`)
		return
	}
	
	mode := os.Args[1]
	
	if mode == "--init"{
		if len(os.Args) < 3 {
			fmt.Println("Usage: go run zkSnark.go --init <domain>")
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
		go_modules.CallBashFiles("./runCircom.sh", TLD_INPUT_FOLDER, "../.././circom_inputs_init/tldCircomInput.json")
		go_modules.CallBashFiles("./runCircom.sh", TARGET_INPUT_FOLDER, "../.././circom_inputs_init/targetCircomInput.json")

		fmt.Println("Finished init zk-snark circom")

	}  else if mode == "--node-generate-prove" {
		if len(os.Args) < 4 {
			fmt.Println("Usage: go run zkSnark.go --node-generate-prove <--tld or --target> <input_json_file_to_be_proved>")
			return
		}
		domain_mode := go_modules.RemovePrefixFlag(os.Args[2])
		prove_file_path := os.Args[3]
		
		input_file_name := go_modules.GetFileNameWithoutExt(prove_file_path)
		prove_folder_name := "../" + input_file_name + "-" + domain_mode

		go_modules.CreateFolder(prove_folder_name)

		// create format binaries json
		go_modules.CheckFormatSpecPy(prove_file_path, prove_folder_name + PROVE_FORMAT_OUTPUT)

		// get the name and id field, convert to int
		go_modules.NameIdConvertPy(prove_file_path, prove_folder_name + PROVE_NAME_ID_OUTPUT)

		// combine both files spec (main & id) + (format binaries) to create circom json input
		go_modules.CombineJSONFile(prove_folder_name + PROVE_FORMAT_OUTPUT, prove_folder_name + PROVE_NAME_ID_OUTPUT, prove_folder_name + PROVE_CIRCOM_INPUT)
		
		// run circom zksnark proofs
		go_modules.CallBashFiles("./runCircom.sh", PROVE_CIRCOM_INPUT_FOLDER + "-" + domain_mode, "../" + prove_folder_name + PROVE_CIRCOM_INPUT)
		
		fmt.Println("Finished creating proves zk-snark circom")

	} else if mode == "--verify-prove"{
		if len(os.Args) < 4 {
			fmt.Println("Usage: go run zkSnark.go --verify-prove <--tld or --target> <prove_folder_to_be_verified>")
			return
		}
		tld_or_target := os.Args[2]
		prove_folder := os.Args[3]

		switch tld_or_target{
			case "--target":
				go_modules.CallBashFiles("./verifyProof.sh", TARGET_INPUT_FOLDER + VERIFICATION_KEY_PATH,
								prove_folder + PUBLIC_JSON_PATH, TARGET_INPUT_FOLDER + PROOF_JSON_PATH)
			case "--tld":
				go_modules.CallBashFiles("./verifyProof.sh", TLD_INPUT_FOLDER + VERIFICATION_KEY_PATH,
								prove_folder + PUBLIC_JSON_PATH, TLD_INPUT_FOLDER + PROOF_JSON_PATH)
			default:
				fmt.Println("Usage: go run main.go --node-prove <--tld or --target> <input_file_to_be_proved>")
				return
		}

	} else if mode == "--cleanup" {
		go_modules.CallBashFiles("./cleanCircomFiles.sh", "tldCircomInput")
		go_modules.CallBashFiles("./cleanCircomFiles.sh", "targetCircomInput")

		fmt.Println("Finished cleaning up")
	} else {
		fmt.Println("Usage: go run initZkSnark.go --init <domain> OR go run initZkSnark.go --cleanup")
	}
}	
