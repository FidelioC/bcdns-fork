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

1.  cd ./circom_zksnark/scripts/ && go run initZkSnark.go --init example.com
2.  go run initZkSnark.go --node-prove ../test/targetChainSpecTest.json
3.  cd ./circom_zksnark/scripts/ && go run initZkSnark.go --cleanup
