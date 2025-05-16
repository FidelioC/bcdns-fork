pragma circom 2.0.0;

include "../circom/node_modules/circomlib/circuits/poseidon.circom";

// to verify the target of the connection string
template source_conn_string() {
    signal input name;
    signal input id;
    signal input bootNodes;
    signal input telemetryEndpoints;
    signal input protocolID;
    signal input properties;
    signal input codeSubstitutes;
    
    signal output hashOut;

    component hash = Poseidon(2);
    hash.inputs[0] <== name;
    hash.inputs[1] <== id;

    hashOut <== hash.out;
}

component main = source_conn_string();