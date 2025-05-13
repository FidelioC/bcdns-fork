pragma circom 2.0.0;

// to check if the connection string format is correct
template check_format_correct(){
    signal input has_name;
    signal input has_id;
    signal input has_chainType;
    signal input has_bootNodes;
    signal input has_telemetryEndpoints;
    signal input has_protocolId;
    signal input has_properties;
    signal input has_codeSubstitutes;

    signal output all_valid;

    // check each value = 1, constraint
    has_name === 1;
    has_id === 1;
    has_chainType === 1;
    has_bootNodes === 1;
    has_telemetryEndpoints === 1;
    has_protocolId === 1;
    has_properties === 1;
    has_codeSubstitutes === 1;

    // format valid, iff all inputs are 1 or total = 8
    all_valid <== has_name + has_id + has_chainType + has_bootNodes + has_telemetryEndpoints + 
            has_protocolId + has_properties + has_codeSubstitutes;
}

// to verify the source of the connection string
template source_conn_string() {
    signal input name;
    signal input id;
    signal input chainType;
    signal input bootNodes;
    signal input telemetryEndpoints;
    signal input protocolId;
    signal input properties;
    signal input codeSubstitutes;
    
}

component main = check_format_correct();