pragma circom 2.0.0;

template Main() {
    // Format check inputs
    signal input has_name;
    signal input has_id;
    signal input has_bootNodes;

    // Hash inputs
    signal input name;
    signal input id;

    // Outputs
    signal output all_valid;
    signal output hashOut;

    // Enforce all are 1 (format check)
    has_name === 1;
    has_id === 1;
    has_bootNodes === 1;

    // Compute format validity (sum should be 8)
    all_valid <== has_name + has_id + has_bootNodes;

    // Hash name and id
    component hash = Poseidon(2);
    hash.inputs[0] <== name;
    hash.inputs[1] <== id;
    hashOut <== hash.out;
}

component main = Main();
