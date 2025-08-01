# Circom Setup

This page will give you a step by step instruction on how to setup circom, used for the zero knowledge proof.

## Circom Installation:

Under circom_zksnark/scripts/, clone the `circom` repository:

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

- circomlib setup:
  npm i circomlib

- snarkjs setup:
  npm install -g snarkjs

- example usage overall:

1.  cd ./circom_zksnark/scripts/
2.  go run main.go --init example.com
3.  go run main.go --node-generate-prove --tld ../test/tldChainSpecTest.json
4.  go run main.go --verify-prove --tld ../proveCircomFiles-tld
5.  go run main.go --cleanup
