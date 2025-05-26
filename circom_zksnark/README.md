- circom setup:

  - git clone https://github.com/iden3/circom.git
  - cd circom
  - cargo build --release
  - cargo install --path circom

- circomlib setup:

  - npm i circomlib

- example usage script by script:

1.  go run initZkSnark.go example.com
2.  ./runCircom.sh tldCircomInput
3.  ./runCircom.sh targetCircomInput
4.  ./cleanFiles.sh tldCircomInput
5.  ./cleanFiles.sh targetCircomInput

- example usage overall:

1.  cd ./circom_zksnark/scripts/
2.  go run zkSnark.go --init example.com
3.  go run zkSnark.go --node-generate-prove --tld ../test/tldChainSpecTest.json
4.  go run zkSnark.go --verify-prove --tld ../proveCircomFiles-tld
5.  go run zkSnark.go --cleanup
