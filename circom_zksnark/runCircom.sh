#!/bin/bash
set -e

cd circomCombined

# compile circom file
circom circomCheck.circom --r1cs --wasm --sym

cd circomCheck_js

# generate witness
node generate_witness.js circomCheck.wasm ../../specInputCombined.json witness.wtns

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