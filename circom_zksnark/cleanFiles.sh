#!/bin/bash
set -e

# Remove circoms
cd circomCombined

# Remove Circom build outputs
rm -f *.r1cs
rm -f *.sym
rm -f *.zkey
rm -f *.ptau
rm -f *.json
rm -f *.wtns

# Remove WASM and JS folders
rm -rf *_js

# Remove json files created from initSpecInput.json
cd ..
rm -rf tldSpecInput.json
rm -rf targetSpecInput.json
rm -rf validSpecResult.json
rm -rf tldCircomInput.json
rm -rf targetCircomInput.json


echo "Cleanup complete!"