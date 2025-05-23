import json
import click

def string_to_base256_int(s):
    result = 0
    base = 256
    for char in s:
        result = result * base + ord(char)
    return result

@click.command()
@click.option('--input', '-i', 'input_path', required=True, help='Path to the input JSON file')
@click.option('--output', '-o', 'output_path', required=True, help='Path to the output JSON file')
def main(input_path, output_path):
    with open(input_path, 'r') as infile:
        data = json.load(infile)

    result = {}
    for key in ['name', 'id']:
        if key in data and isinstance(data[key], str):
            result[key] = string_to_base256_int(data[key])

    with open(output_path, 'w') as outfile:
        json.dump(result, outfile, indent=2)

    click.echo(f"Converted fields written to {output_path}")

if __name__ == '__main__':
    main()
