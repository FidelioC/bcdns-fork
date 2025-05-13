import json
import click

# usage: python3 key_presence.py --input-file rootSpec.json --output-file conn_string_result.json

required_keys = [
    "name",
    "id",
    "chainType",
    "bootNodes",
    "telemetryEndpoints",
    "protocolId",
    "properties",
    "codeSubstitutes",
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

    # print(result)


if __name__ == "__main__":
    main()
