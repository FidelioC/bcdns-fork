package scripts

import (
	"fmt"
)

const (
	INIT_FOLDER_NAME          = "../circom_inputs_init"
	TLD_FILE_NAME             = "../circom_inputs_init/tldSpecInput.json"
	TARGET_FILE_NAME          = "../circom_inputs_init/targetSpecInput.json"
	FORMAT_RESULT_INPUT       = "../config/validChainSpec.json"
	FORMAT_RESULT_OUTPUT      = "../circom_inputs_init/validSpecResult.json"
	TLD_CIRCOM_INPUT          = "../circom_inputs_init/tldCircomInput.json"
	TLD_INPUT_FOLDER          = "../tldCircomInput"
	TARGET_CIRCOM_INPUT       = "../circom_inputs_init/targetCircomInput.json"
	TARGET_INPUT_FOLDER       = "../targetCircomInput"
	PROVE_CIRCOM_INPUT_FOLDER = "../proveCircomFiles"
	PROVE_FORMAT_OUTPUT       = "/proveSpecResult.json"
	PROVE_NAME_ID_OUTPUT      = "/proveNameIdOutput.json"
	PROVE_CIRCOM_INPUT        = "/proveCircomInput.json"
	VERIFICATION_KEY_PATH     = "/circomCheck_js/verification_key.json"
	PUBLIC_JSON_PATH          = "/circomCheck_js/public.json"
	PROOF_JSON_PATH           = "/circomCheck_js/proof.json"
)

func InitProve(domain string) {
	CreateFolder(INIT_FOLDER_NAME)

	// create tld and target spec input json file based off the domain name
	InitSpecInputPy("initSpecInput.py", domain, TLD_FILE_NAME, TARGET_FILE_NAME)

	// create format binaries json
	CheckFormatSpecPy("checkSpecFormat.py", FORMAT_RESULT_INPUT, FORMAT_RESULT_OUTPUT)

	// combine both files spec (main & id) + (format binaries) to create circom json input
	CombineJSONFile("combineJson.py", TLD_FILE_NAME, FORMAT_RESULT_OUTPUT, TLD_CIRCOM_INPUT)
	CombineJSONFile("combineJson.py", TARGET_FILE_NAME, FORMAT_RESULT_OUTPUT, TARGET_CIRCOM_INPUT)

	// create circom prove
	CallBashFiles("./runCircom.sh", TLD_INPUT_FOLDER, "../.././circom_inputs_init/tldCircomInput.json")
	CallBashFiles("./runCircom.sh", TARGET_INPUT_FOLDER, "../.././circom_inputs_init/targetCircomInput.json")

	fmt.Println("Finished init zk-snark circom")
}

func GenerateProve(domain_mode string, prove_file_path string) {
	input_file_name := GetFileNameWithoutExt(prove_file_path)
	prove_folder_name := "../" + input_file_name + "-" + domain_mode

	CreateFolder(prove_folder_name)

	// create format binaries json
	CheckFormatSpecPy("checkSpecFormat.py", prove_file_path, prove_folder_name+PROVE_FORMAT_OUTPUT)

	// get the name and id field, convert to int
	NameIdConvertPy("nameIdIntConvert.py", prove_file_path, prove_folder_name+PROVE_NAME_ID_OUTPUT)

	// combine both files spec (main & id) + (format binaries) to create circom json input
	CombineJSONFile("combineJson.py", prove_folder_name+PROVE_FORMAT_OUTPUT, prove_folder_name+PROVE_NAME_ID_OUTPUT, prove_folder_name+PROVE_CIRCOM_INPUT)

	// run circom zksnark proofs
	CallBashFiles("./runCircom.sh", PROVE_CIRCOM_INPUT_FOLDER+"-"+domain_mode, "../"+prove_folder_name+PROVE_CIRCOM_INPUT)

	fmt.Println("Finished creating proves zk-snark circom")
}

func VerifyProve(tld_or_target string, prove_folder string) {
	switch tld_or_target {
	case "--target":
		CallBashFiles("./verifyProof.sh", TARGET_INPUT_FOLDER+VERIFICATION_KEY_PATH,
			prove_folder+PUBLIC_JSON_PATH, TARGET_INPUT_FOLDER+PROOF_JSON_PATH)
	case "--tld":
		CallBashFiles("./verifyProof.sh", TLD_INPUT_FOLDER+VERIFICATION_KEY_PATH,
			prove_folder+PUBLIC_JSON_PATH, TLD_INPUT_FOLDER+PROOF_JSON_PATH)
	default:
		fmt.Println("Usage: go run main.go --node-prove <--tld or --target> <input_file_to_be_proved>")
		return
	}

	fmt.Println("Finished verifying prove")
}

func Cleanup() {
	CallBashFiles("./cleanCircomFiles.sh", "tldCircomInput")
	CallBashFiles("./cleanCircomFiles.sh", "targetCircomInput")

	fmt.Println("Finished cleaning up")
}
