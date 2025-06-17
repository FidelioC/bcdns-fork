# DNS Client

This package provides an implementation of a DNS client for the decentralized architecture presented in this project.

---

## Installation

### Local Installation

1. Install dependencies:
   ```bash
   npm install
   ```

2. Configure the network:
   Update the `ROOT_DNS_NETWORK_SPEC_ADDR` in `config.js` with the address of your root DNS network.

### Docker Installation

1. Build the Docker image:
   ```bash
   docker-compose build
   ```

2. Configure the network:
   Update the `ROOT_DNS_NETWORK_SPEC_ADDR` in `config.js` with the address of your root DNS network.

---

## Usage

### 1. Resolving a Domain

To resolve a domain, use the following command:

```bash
npm start <domain>
```

#### Example:
```bash
npm start example.tld
```

#### Output:
- If the domain has a `targetSpec`, it will display the target specification.
- If the domain represents an asset, it will display the asset details.

---

### 2. Operation Modes

The DNS client supports three operation modes:

#### A. Web Interface Mode

##### Local Usage:
```bash
BCDNS_MODE=web npm start
```

##### Docker Usage:
```bash
docker compose run bcdns-web
```
Access the web interface at http://localhost:3000

#### B. Interactive CLI Mode

##### Local Usage:
```bash
BCDNS_MODE=cli npm start
```

##### Docker Usage:
```bash
docker compose run bcdns-cli
```

Available commands in CLI mode:
- `domain <domainName>` - Resolve a domain
- `asset <assetName>` - Resolve an asset
- `exit` - Quit the application

#### C. One-time Resolution Mode

##### Local Usage:
```bash
# Resolve a domain
npm start -- domain example.com

# Resolve an asset
npm start -- asset example.com/123
```

##### Docker Usage:
```bash
# Resolve a domain
docker compose run bcdns-onetime domain example.com

# Resolve an asset
docker compose run bcdns-onetime asset example.com/123
```

### 3. Registering a TLD, Domain, or Asset

To register a TLD, domain, or asset, use the `npm run register` command with the appropriate flags:

#### Register a TLD:
```bash
# Local
npm run register --tld <tld> <spec> <phrase...>

# Docker
docker compose run bcdns-register npm run register -- --tld <tld> <spec> <phrase...>
```
- `<tld>`: The top-level domain to register.
- `<spec>`: The specification for the TLD.
- `<phrase...>`: Optional phrases for additional information.

##### Example:
```bash
# Local
npm run register --tld example "TLD specification" "TLD description"

# Docker
docker compose run bcdns-register npm run register -- --tld example "TLD specification" "TLD description"
```

#### Register a Domain:
```bash
# Local
npm run register --domain <domain> <spec> <phrase...>

# Docker
docker compose run bcdns-register npm run register -- --domain <domain> <spec> <phrase...>
```
- `<domain>`: The domain to register.
- `<spec>`: The specification for the domain.
- `<phrase...>`: Optional phrases for additional information.

##### Example:
```bash
# Local
npm run register --domain sub.example.tld "Domain specification" "Domain description"

# Docker
docker compose run bcdns-register npm run register -- --domain sub.example.tld "Domain specification" "Domain description"
```

#### Register an Asset:
```bash
# Local
npm run register --asset <domain> <assetId> <amount>

# Docker
docker compose run bcdns-register npm run register -- --asset <domain> <assetId> <amount>
```
- `<domain>`: The domain associated with the asset.
- `<assetId>`: The identifier for the asset.
- `<amount>`: The amount of the asset.

##### Example:
```bash
# Local
npm run register --asset example.tld asset123 100

# Docker
docker compose run bcdns-register npm run register -- --asset example.tld asset123 100
```

#### Help Command:
To display usage instructions for registration:
```bash
# Local
npm run register --help

# Docker
docker compose run bcdns-register npm run register -- --help
```

