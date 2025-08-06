import json
import click

# ============================================
# Script Name : checkSpecFormat.py
# Author      : Fidelio Ciandy
# Description : Checks the Chain Spec JSON format and outputs a result summary.
# Usage       : python3 checkSpecFormat.py --input-file rootSpec.json --output-file conn_string_result.json
#
# Example Input (rootSpec.json):
# {
#   "name": "root",
#   "id": "root",
#   "bootNodes": [
#     "/ip4/172.18.0.2/tcp/9945/p2p/12D3KooWNL4mZo8y7oAes3VRRnbHy91TDLxnjrDsnMFZkPebB2Rh",
#     "/ip4/172.18.0.3/tcp/9945/p2p/12D3KooWNL4mZo8y7oAes3VRRnbHy91TDLxnjrDsnMFZkPebB2Rh"
#   ]
# }
#
# Example Output (conn_string_result.json):
# {
#   "name": 1,
#   "id": 1,
#   "bootNodes": 1
# }
# ============================================


required_keys = [
    "name",
    "id",
    "bootNodes",
]

result_dict = {}


def key_presence(required_keys, source_json):
    for key in required_keys:
        result_dict[f"has_{key}"] = 1 if key in source_json else 0
    return result_dict


@click.command()
@click.option(
    "--input-file",
    required=True,
    type=click.Path(exists=True),
    help="Path to the input JSON file",
)
@click.option(
    "--output-file",
    required=True,
    type=click.Path(),
    help="Path to save the result JSON file",
)
def main(input_file, output_file):
    with open(input_file, "r") as conn_str:
        connection_string = json.load(conn_str)

    result = key_presence(required_keys, connection_string)

    with open(output_file, "w") as result_json:
        json.dump(result, result_json, indent=2)

    print(f"checkSpecFormat.py: {output_file} created successfully")


if __name__ == "__main__":
    main()
