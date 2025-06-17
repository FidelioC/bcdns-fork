const { createResolver } = require('./dns/resolver');
const { ROOT_DNS_NETWORK_SPEC_ADDR } = require('./config');
const { WebServer } = require('./application/web/server');
const { CLIInterface } = require('./application/cli/interface');

const MODE = process.env.BCDNS_MODE || 'onetime';
const WEB_PORT = parseInt(process.env.BCDNS_WEB_PORT || '3000');

const runOneTime = async () => {
    console.log(process.argv.length)
    if (process.argv.length < 4) {
        console.error('Usage:\n\nnpm start -- <type> <value>\nwhere type is either "domain" or "asset"');
        process.exit(1);
    }

    const type = process.argv[2].toLowerCase();
    const value = process.argv[3];
    
    const resolver = createResolver(ROOT_DNS_NETWORK_SPEC_ADDR);
    await resolver.init();

    let result;
    if (type === 'domain') {
        result = await resolver.resolve(value);
    } else if (type === 'asset') {
        result = await resolver.resolveAsset(value);
    } else {
        console.error('Type must be either "domain" or "asset"');
        process.exit(1);
    }

    if (result.targetSpec) {
        console.log('================================================================');
        console.log(`               TARGET SPEC FOUND FOR ${value}`);
        console.log('================================================================');
        console.log(result.targetSpec);
        console.log('****************************************************************');
    } else if (result.asset) {
        console.log('================================================================');
        console.log(`               ASSET FOUND FOR ${value}`);
        console.log('================================================================');
        console.log(result.asset);
        console.log('****************************************************************');
    }
};

const runCLI = async () => {
    const cli = new CLIInterface();
    await cli.init();
    await cli.start();
};

const runWeb = async () => {
    const server = new WebServer();
    await server.init();
    server.start(WEB_PORT);
};

const main = async () => {
    try {
        switch (MODE) {
            case 'onetime':
                await runOneTime();
                break;
            case 'cli':
                await runCLI();
                break;
            case 'web':
                await runWeb();
                break;
            default:
                console.error('Invalid BCDNS_MODE. Must be one of: onetime, cli, web');
                process.exit(1);
        }
    } catch (error) {
        console.error('Error:', error.message);
        process.exit(1);
    }
};

main();