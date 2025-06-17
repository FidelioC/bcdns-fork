const readline = require('readline');
const { createResolver } = require('../../dns/resolver');
const { ROOT_DNS_NETWORK_SPEC_ADDR } = require('../../config');

class CLIInterface {
    constructor() {
        this.resolver = createResolver(ROOT_DNS_NETWORK_SPEC_ADDR);
        this.rl = readline.createInterface({
            input: process.stdin,
            output: process.stdout
        });
    }

    async init() {
        await this.resolver.init();
    }

    printResult(result) {
        if (result.targetSpec) {
            console.log("================================================================");
            console.log("               TARGET SPEC FOUND");
            console.log("================================================================");
            console.log(result.targetSpec);
            console.log("****************************************************************");
        } else if (result.asset) {
            console.log("================================================================");
            console.log("               ASSET FOUND");
            console.log("================================================================");
            console.log(result.asset);
            console.log("****************************************************************");
        } else {
            console.log("No result found");
        }
    }

    async start() {
        console.log('BCDNS CLI Interface');
        console.log('Commands:');
        console.log('  domain <domainName> - resolve domain');
        console.log('  asset <assetName> - resolve asset');
        console.log('  exit - quit the application');
        console.log('');

        while (true) {
            const command = await new Promise(resolve => {
                this.rl.question('> ', resolve);
            });

            if (command.toLowerCase() === 'exit') {
                this.rl.close();
                process.exit(0);
            }

            const [cmd, ...args] = command.split(' ');
            const input = args.join(' ');

            try {
                switch (cmd.toLowerCase()) {
                    case 'domain':
                        if (!input) {
                            console.log('Please provide a domain name');
                            continue;
                        }
                        this.printResult(await this.resolver.resolve(input));
                        break;
                    case 'asset':
                        if (!input) {
                            console.log('Please provide an asset name');
                            continue;
                        }
                        // Format the input as domain.tld/asset/assetId
                        const parts = input.split('/');
                        if (parts.length !== 2) {
                            console.log('Asset format should be: domain.tld/assetId');
                            continue;
                        }
                        const formattedInput = `${parts[0]}/asset/${parts[1]}`;
                        this.printResult(await this.resolver.resolve(formattedInput));
                        break;
                    default:
                        console.log('Unknown command. Available commands: domain, asset, exit');
                }
            } catch (error) {
                console.error('Error:', error.message);
            }
        }
    }
}

module.exports = { CLIInterface };
