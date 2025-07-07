pragma circom 2.0.0;

include "./node_modules/circomlib/circuits/poseidon.circom";

template Main() {
    // Format check inputs
    signal input has_name;
    signal input has_id;
    signal input has_chainType;
    signal input has_bootNodes;
    signal input has_telemetryEndpoints;
    signal input has_protocolId;
    signal input has_properties;
    signal input has_codeSubstitutes;

    // Hash inputs
    signal input name;
    signal input id;

    // Outputs
    signal output all_valid;
    signal output hashOut;

    // Enforce all are 1 (format check)
    has_name === 1;
    has_id === 1;
    has_chainType === 1;
    has_bootNodes === 1;
    has_telemetryEndpoints === 1;
    has_protocolId === 1;
    has_properties === 1;
    has_codeSubstitutes === 1;

    // Compute format validity (sum should be 8)
    all_valid <== has_name + has_id + has_chainType + has_bootNodes +
                  has_telemetryEndpoints + has_protocolId + has_properties +
                  has_codeSubstitutes;

    // Hash name and id
    component hash = Poseidon(2);
    hash.inputs[0] <== name;
    hash.inputs[1] <== id;
    hashOut <== hash.out;
}

component main = Main();
