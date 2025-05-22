#!/bin/bash
set -e

folder_path=$1 # new folder path for circom files
input_file_loc=$2 # location of spec input file

# make folder for the desired input json
mkdir $folder_path

cd $folder_path 

# compile circom file
circom ../scripts/circomCheck.circom --r1cs --wasm --sym

cd circomCheck_js

# generate witness
node generate_witness.js circomCheck.wasm $input_file_loc witness.wtns

# powers of tau
snarkjs powersoftau new bn128 12 pot12_0000.ptau -v
snarkjs powersoftau contribute pot12_0000.ptau pot12_0001.ptau --name="First contribution" -v
snarkjs powersoftau prepare phase2 pot12_0001.ptau pot12_final.ptau -v

# contribute
snarkjs groth16 setup ../circomCheck.r1cs pot12_final.ptau circomCheck_0000.zkey
snarkjs zkey contribute circomCheck_0000.zkey circomCheck_0001.zkey --name="1st Contributor Name" -v
snarkjs zkey export verificationkey circomCheck_0001.zkey verification_key.json

# prove
snarkjs groth16 prove circomCheck_0001.zkey ./witness.wtns proof.json public.json

# verify
snarkjs groth16 verify verification_key.json public.json proof.json