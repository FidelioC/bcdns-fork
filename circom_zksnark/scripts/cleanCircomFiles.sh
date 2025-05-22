#!/bin/bash
set -e

input_file=$1 # input file name

# Remove circoms
rm -rf ../$input_file

# Remove json files created from initSpecInput.json
rm -rf ../circom_inputs_init

# Remove circoms prove
rm -rf ../circom_proves
rm -rf ../proveCircomFiles

echo "Cleanup complete!"