import json
import click
import sys

# ============================================
# Script Name : combineJson.py
# Author      : Fidelio Ciandy
# Description : Combine two json files two one singular file
# Usage       : python3 combineJson.py --file1 file1.json --file2 file2.json --output output_file.json
# ============================================


def load_json_file(filepath):
    try:
        with open(filepath, "r") as f:
            return json.load(f)
    except Exception as e:
        click.echo(f"Error loading {filepath}: {e}")
        sys.exit(1)


def merge_json(data1, data2):
    return {**data1, **data2}


def write_json_file(filepath, data):
    try:
        with open(filepath, "w") as f:
            json.dump(data, f, indent=2)
        click.echo(
            f"combineJson.py: Combined JSON written to {filepath} created successfully"
        )
    except Exception as e:
        click.echo(f"Error writing {filepath}: {e}")
        sys.exit(1)


@click.command()
@click.option("--file1", required=True, type=click.Path(exists=True))
@click.option("--file2", required=True, type=click.Path(exists=True))
@click.option("--output", required=True, type=click.Path())
def main(file1, file2, output):
    """Combine two JSON files"""
    data1 = load_json_file(file1)
    data2 = load_json_file(file2)
    combined = merge_json(data1, data2)
    write_json_file(output, combined)


if __name__ == "__main__":
    main()
