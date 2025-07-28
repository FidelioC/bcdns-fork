package main

import (
	"fmt"
	"log"
	"os"

	"github.com/FidelioC/circom_zksnark/scripts"
)

func main(){
	err := os.Chdir("./scripts")
	if err != nil {
		log.Fatalf("Failed to change directory: %v", err)
	}
	
	if len (os.Args) < 2{
		fmt.Println(`Usages: 
			- go run main.go --init <domain> 
			- go run main.go --node-generate-prove <input_json_file_to_be_proved>
			- go run main.go --verify-prove <--tld or --target> <prove_folder_to_be_verified>
			- go run main.go --cleanup`)
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
		scripts.InitProve(domain)

	}  else if mode == "--node-generate-prove" {
		if len(os.Args) < 4 {
			fmt.Println("Usage: go run main.go --node-generate-prove <--tld or --target> <input_json_file_to_be_proved>")
			return
		}
		domain_mode := scripts.RemovePrefixFlag(os.Args[2])
		prove_file_path := os.Args[3]
		scripts.GenerateProve(domain_mode, prove_file_path)

	} else if mode == "--verify-prove"{
		if len(os.Args) < 4 {
			fmt.Println("Usage: go run main.go --verify-prove <--tld or --target> <prove_folder_to_be_verified>")
			return
		}
		tld_or_target := os.Args[2]
		prove_folder := os.Args[3]

		scripts.VerifyProve(tld_or_target, prove_folder)
	} else if mode == "--cleanup" {
		scripts.Cleanup()
	} else {
		fmt.Println("Usage: go run main.go --init <domain> OR go run main.go --cleanup")
	}
}	
