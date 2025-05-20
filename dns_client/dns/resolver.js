const {
    getJSONResponse,
    getTargetSpec,
    getTLDSpec,
    getTLD,
    connector
} = require("../util");

/**
 * Resolves blockchain domains and assets using a root network's chain specification.
 */
class DNSResolver {
    /**
     * @param {string} rootNetworkSpecUrl - URL of the root network's chain specification.
     */
    constructor(rootNetworkSpecUrl) {
        this.rootNetworkSpecUrl = rootNetworkSpecUrl;
        this.rootSpec = null; // Holds the fetched root network's chain specification
    }

    /**
     * Initializes the resolver by loading the root network's chain specification.
     * @returns {Promise<void>}
     * @throws {Error} If the root specification cannot be fetched.
     */
    async init() {
        this.rootSpec = await getJSONResponse(this.rootNetworkSpecUrl);
    }

    /**
     * Resolves details of an asset on a blockchain.
     * @param {object} chainSpec - The chain specification of the target network.
     * @param {string} assetId - Unique identifier of the asset.
     * @returns {Promise<object>} Human-readable asset details.
     * @throws {Error} If the asset query or network connection fails.
     */
    async resolveAsset(chainSpec, assetId) {
        let api = await connector.connectToNetwork(chainSpec);
        let asset = await api.query.assetsModule.asset(assetId);
        return asset.toHuman();
    }

    /**
     * Resolves a domain to its blockchain network or asset details.
     * @param {string} domain - The domain to resolve, optionally including an asset ID.
     * @returns {Promise<object>} Resolution result containing asset or target network details.
     * @throws {Error} If domain parsing or resolution fails.
     */
    async resolve(domain) {
        let parsedDomain = this.#parseAssetDomain(domain);
        console.log("parsedDomain " + parsedDomain.domain)
        let tld = getTLD(parsedDomain.domain);
        console.log("tld " + tld)
        let tldSpec = await getTLDSpec(tld, this.rootSpec);
        console.log("tldSpec " + JSON.stringify(tldSpec))
        let targetSpec = await getTargetSpec(parsedDomain.domain, tldSpec);
        console.log("targetSpec " + JSON.stringify(targetSpec))

        if (parsedDomain.assetId) {
            let asset = await this.resolveAsset(targetSpec, parsedDomain.assetId);
            return { asset };
        } else {
            return { targetSpec };
        }
    }

    /**
     * Parses a domain string to extract the base domain and optional asset ID.
     * @param {string} domain - Domain string in the format "domain.tld/asset/assetId".
     * @returns {object} Parsed domain and asset ID (if present).
     */
    #parseAssetDomain(domain) {
        let domainArr = domain.split('/');
        return {
            domain: domainArr[0],
            assetId: domainArr.length < 3 || domainArr[1] !== 'asset' ? null : domainArr[2]
        };
    }
}

/**
 * Factory function to create a new DNSResolver instance.
 * @param {string} rootSpecAddr - URL of the root network's chain specification.
 * @returns {DNSResolver} A new DNSResolver instance.
 */
exports.createResolver = (rootSpecAddr) => {
    return new DNSResolver(rootSpecAddr);
};
