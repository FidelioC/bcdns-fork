# usage example: ./verifyProof.sh ../targetCircomInput/circomCheck_js/verification_key.json ../proveCircomFiles/circomCheck_js/public.json ../targetCircomInput/circomCheck_js/proof.json

#!/bin/bash
set -e

verification_key_file=$1
public_filepath=$2
proof_filepath=$3

# verify
snarkjs groth16 verify $verification_key_file $public_filepath $proof_filepath

