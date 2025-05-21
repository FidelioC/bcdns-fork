import click
import json

class SpecInput:
    def __init__(self, name, id):
        self.name = name
        self.id = id

    def to_dict(self):
        return {
            "name": self.name,
            "id": self.id
        }

def split_domain(domain_input):
    split = domain_input.split('.')
    if len(split) != 2:
        raise ValueError(f"Invalid domain format: {domain_input}")
    return split[0], split[1]

def string_to_number(string_input):
    # Base-256 encoding
    result = 0
    for c in string_input:
        result = result * 256 + ord(c)
    return result

def create_spec_json(tld_input):
    encoded = string_to_number(tld_input)
    return SpecInput(encoded, encoded)

def create_spec_json_file(spec_input, file_name):
    try:
        with open(file_name, 'w') as f:
            json.dump(spec_input.to_dict(), f, indent=2)
        print(f"initSpecInput.py: {file_name} created successfully")
    except Exception as e:
        print(f"Error creating {file_name}: {e}")

def create_json_file(string_input, file_name):
    spec_input = create_spec_json(string_input)
    create_spec_json_file(spec_input, file_name)

@click.command()
@click.option('--domain', required=True, help='Domain name (e.g., example.com)')
@click.option('--tld-output', required=True, help='Output JSON file for the TLD')
@click.option('--target-output', required=True, help='Output JSON file for the domain (target) name')
def main(domain, tld_output, target_output):
    try:
        domain_name, tld = split_domain(domain)
    except ValueError as e:
        print(f"Error: {e}")
        return
    
    # print(f"\nInput: {domain}")
    # print(f"Domain name: {domain_name}")
    # print(f"TLD: {tld}")

    create_json_file(tld, tld_output)
    create_json_file(domain_name, target_output)

if __name__ == '__main__':
    main()
