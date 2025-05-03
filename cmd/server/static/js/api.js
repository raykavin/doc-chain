/**
 * DocChain - API Client
 * Handles communication with the backend API
 */

const API_BASE_URL = '/api';

// API endpoints
const ENDPOINTS = {
    // Wallet endpoints
    WALLET: {
        CREATE: `${API_BASE_URL}/wallet/new`,
        GET: (id) => `${API_BASE_URL}/wallet/${id}`,
        LIST: `${API_BASE_URL}/wallet/list`
    },
    
    // Document endpoints
    DOCUMENT: {
        SIGN: `${API_BASE_URL}/document/sign`,
        GET: (id) => `${API_BASE_URL}/document/${id}`,
        VERIFY: (id) => `${API_BASE_URL}/document/verify/${id}`,
        LIST: `${API_BASE_URL}/document/list`,
        UPLOAD: `${API_BASE_URL}/document/upload`
    },
    
    // Blockchain endpoints
    BLOCKCHAIN: {
        INFO: `${API_BASE_URL}/blockchain`,
        MINE: `${API_BASE_URL}/blockchain/mine`,
        BLOCKS: `${API_BASE_URL}/blockchain/blocks`,
        BLOCK: (hash) => `${API_BASE_URL}/blockchain/block/${hash}`,
        PENDING: `${API_BASE_URL}/blockchain/pending`,
        VALIDATE: `${API_BASE_URL}/blockchain/validate`
    },
    
    // ICP-Brasil endpoints
    ICPBRASIL: {
        SIGN: `${API_BASE_URL}/icpbrasil/sign`,
        VERIFY: `${API_BASE_URL}/icpbrasil/verify`,
        CERT_INFO: `${API_BASE_URL}/icpbrasil/cert-info`,
        CHECK_REVOCATION: `${API_BASE_URL}/icpbrasil/check-revocation`,
        DOCUMENT: (id) => `${API_BASE_URL}/icpbrasil/document/${id}`
    },
    
    // Template endpoints
    TEMPLATE: {
        LIST: `${API_BASE_URL}/template/list`,
        GET: (id) => `${API_BASE_URL}/template/${id}`,
        CREATE: `${API_BASE_URL}/template/create`,
        UPDATE: (id) => `${API_BASE_URL}/template/${id}`,
        DELETE: (id) => `${API_BASE_URL}/template/${id}`,
        USE: `${API_BASE_URL}/template/use`
    }
};

/**
 * API Client for DocChain
 */
class DocChainAPI {
    /**
     * Make a request to the API
     * @param {string} url - The URL to request
     * @param {string} method - The HTTP method to use
     * @param {object} data - The data to send (for POST/PUT requests)
     * @returns {Promise<object>} - The response data
     */
    async request(url, method = 'GET', data = null) {
        const options = {
            method,
            headers: {
                'Content-Type': 'application/json',
                'Accept': 'application/json'
            }
        };
        
        if (data && (method === 'POST' || method === 'PUT')) {
            options.body = JSON.stringify(data);
        }
        
        try {
            const response = await fetch(url, options);
            
            if (!response.ok) {
                const errorData = await response.json();
                throw new Error(errorData.message || 'API request failed');
            }
            
            return await response.json();
        } catch (error) {
            console.error('API request error:', error);
            throw error;
        }
    }
    
    // ===== Wallet API =====
    
    /**
     * Create a new wallet
     * @param {string} name - The wallet name
     * @returns {Promise<object>} - The created wallet
     */
    async createWallet(name) {
        return this.request(ENDPOINTS.WALLET.CREATE, 'POST', { name });
    }
    
    /**
     * Get a wallet by ID
     * @param {string} id - The wallet ID
     * @returns {Promise<object>} - The wallet
     */
    async getWallet(id) {
        return this.request(ENDPOINTS.WALLET.GET(id));
    }
    
    // ===== Document API =====
    
    /**
     * Sign a document
     * @param {object} data - The document data
     * @returns {Promise<object>} - The signed document
     */
    async signDocument(data) {
        return this.request(ENDPOINTS.DOCUMENT.SIGN, 'POST', data);
    }
    
    /**
     * Get a document by ID
     * @param {string} id - The document ID
     * @returns {Promise<object>} - The document
     */
    async getDocument(id) {
        return this.request(ENDPOINTS.DOCUMENT.GET(id));
    }
    
    /**
     * Verify a document
     * @param {string} id - The document ID
     * @returns {Promise<object>} - The verification result
     */
    async verifyDocument(id) {
        return this.request(ENDPOINTS.DOCUMENT.VERIFY(id));
    }
    
    // ===== Blockchain API =====
    
    /**
     * Get blockchain information
     * @returns {Promise<object>} - The blockchain info
     */
    async getBlockchainInfo() {
        return this.request(ENDPOINTS.BLOCKCHAIN.INFO);
    }
    
    /**
     * Mine a new block
     * @returns {Promise<object>} - The mined block
     */
    async mineBlock() {
        return this.request(ENDPOINTS.BLOCKCHAIN.MINE, 'POST');
    }
    
    /**
     * Get all blocks
     * @returns {Promise<Array>} - The blocks
     */
    async getBlocks() {
        return this.request(ENDPOINTS.BLOCKCHAIN.BLOCKS);
    }
    
    /**
     * Get a block by hash
     * @param {string} hash - The block hash
     * @returns {Promise<object>} - The block
     */
    async getBlock(hash) {
        return this.request(ENDPOINTS.BLOCKCHAIN.BLOCK(hash));
    }
    
    /**
     * Get pending documents
     * @returns {Promise<Array>} - The pending documents
     */
    async getPendingDocuments() {
        return this.request(ENDPOINTS.BLOCKCHAIN.PENDING);
    }
    
    /**
     * Validate the blockchain
     * @returns {Promise<object>} - The validation result
     */
    async validateChain() {
        return this.request(ENDPOINTS.BLOCKCHAIN.VALIDATE);
    }
    
    // ===== ICP-Brasil API =====
    
    /**
     * Sign a document with ICP-Brasil certificate
     * @param {object} data - The document data
     * @returns {Promise<object>} - The signed document
     */
    async signDocumentWithICPBrasil(data) {
        return this.request(ENDPOINTS.ICPBRASIL.SIGN, 'POST', data);
    }
    
    /**
     * Verify a document signature
     * @param {object} data - The verification data
     * @returns {Promise<object>} - The verification result
     */
    async verifySignature(data) {
        return this.request(ENDPOINTS.ICPBRASIL.VERIFY, 'POST', data);
    }
    
    /**
     * Get certificate information
     * @param {object} data - The certificate data
     * @returns {Promise<object>} - The certificate info
     */
    async getCertificateInfo(data) {
        return this.request(ENDPOINTS.ICPBRASIL.CERT_INFO, 'POST', data);
    }
    
    /**
     * Check certificate revocation
     * @param {object} data - The certificate data
     * @returns {Promise<object>} - The revocation check result
     */
    async checkRevocation(data) {
        return this.request(ENDPOINTS.ICPBRASIL.CHECK_REVOCATION, 'POST', data);
    }
    
    /**
     * Get document information
     * @param {string} id - The document ID
     * @returns {Promise<object>} - The document info
     */
    async getDocumentInfo(id) {
        return this.request(ENDPOINTS.ICPBRASIL.DOCUMENT(id));
    }
    
    // ===== Template API =====
    
    /**
     * Get all templates
     * @returns {Promise<Array>} - The templates
     */
    async getTemplates() {
        return this.request(ENDPOINTS.TEMPLATE.LIST);
    }
    
    /**
     * Get a template by ID
     * @param {string} id - The template ID
     * @returns {Promise<object>} - The template
     */
    async getTemplate(id) {
        return this.request(ENDPOINTS.TEMPLATE.GET(id));
    }
    
    /**
     * Create a new template
     * @param {object} data - The template data
     * @returns {Promise<object>} - The created template
     */
    async createTemplate(data) {
        return this.request(ENDPOINTS.TEMPLATE.CREATE, 'POST', data);
    }
    
    /**
     * Update a template
     * @param {string} id - The template ID
     * @param {object} data - The template data
     * @returns {Promise<object>} - The updated template
     */
    async updateTemplate(id, data) {
        return this.request(ENDPOINTS.TEMPLATE.UPDATE(id), 'PUT', data);
    }
    
    /**
     * Delete a template
     * @param {string} id - The template ID
     * @returns {Promise<object>} - The deletion result
     */
    async deleteTemplate(id) {
        return this.request(ENDPOINTS.TEMPLATE.DELETE(id), 'DELETE');
    }
    
    /**
     * Create a document from a template
     * @param {object} data - The template usage data
     * @returns {Promise<object>} - The created document
     */
    async useTemplate(data) {
        return this.request(ENDPOINTS.TEMPLATE.USE, 'POST', data);
    }
}

// Create a singleton instance
const api = new DocChainAPI();
