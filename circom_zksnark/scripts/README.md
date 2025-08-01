# Circom Setup

This page will give you a step by step instruction on how to setup circom, used for running zero knowledge proof.

## Circom Installation:

Under `circom_zksnark/scripts/`, clone the `circom` repository:

```bash
git clone https://github.com/iden3/circom.git
```

Change directory to the downloaded `circom` folder:

```bash
cd circom
```

Use cargo build to compile:

```bash
cargo build --release
```

Lastly set the circom to known path:

```bash
cargo install --path circom
```

## Installing snarkjs

```bash
npm install -g snarkjs
```

## Installing circomlib

```bash
npm i circomlib
```

## Reference

- Reference for circom and snarkjs installation: [Installing The Circom Ecosystem](https://docs.circom.io/getting-started/installation/#installing-snarkjs)

## Example usage:

- These commands will show you how to run the zero knowledge proof without having to run the whole blockchain discovery process in dns_client_golang/main.go script.

1. Change directory to `circom_zksnark`.

```bash
cd ./circom_zksnark
```

2. Run the initial setup process to generate the expected prove/ certificate files for the desired domain (i.e, this is the verifier).

```bash
go run main.go --init example.com
```

3. Generate prove based on the given input json file (i.e, this is the prover).

```bash
go run main.go --node-generate-prove --tld ../test/tldChainSpecTest.json
```

**Note:** when being asked to "Enter a random text. (Entropy):" in the terimnal, just type in any character and press enter.

4. This process will compare the result from the initial process (2nd step) to the result of the node generate prove process (3rd step) and see if the result from the (3rd step) matches the certificate that has been generated on the 2nd step (i.e, this process compare the prover's result to the verifier's certificate).

```bash
go run main.go --verify-prove --tld ../proveCircomFiles-tld
```

5. Clean up all processing files from previous steps.

```bash
go run main.go --cleanup
```
