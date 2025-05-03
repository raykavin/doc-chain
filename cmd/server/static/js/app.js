/**
 * DocChain - Blockchain Document Signing with ICP-Brasil Support
 * Main application JavaScript
 */

// Mock data for demonstration
const mockData = {
    wallets: [
        { id: '1a2b3c', name: 'Personal Wallet', hasICPCert: true, certType: 'e-CPF' },
        { id: '4d5e6f', name: 'Corporate Wallet', hasICPCert: true, certType: 'e-CNPJ' }
    ],
    documents: [
        { id: 'doc1', title: 'Contract with Client A', status: 'signed', signedBy: '1a2b3c', signatureTime: '2025-05-01T14:30:00Z', contentType: 'pdf' },
        { id: 'doc2', title: 'Invoice #12345', status: 'pending', contentType: 'html' },
        { id: 'doc3', title: 'Terms of Service', status: 'signed', signedBy: '4d5e6f', signatureTime: '2025-04-28T10:15:00Z', contentType: 'md' }
    ],
    templates: [
        { id: 'temp1', name: 'Service Contract', description: 'Standard service agreement', type: 'md' },
        { id: 'temp2', name: 'Invoice', description: 'Basic invoice template', type: 'html' }
    ],
    blocks: [
        { index: 0, hash: '0000abc123def456', documentCount: 1, timestamp: '2025-01-01T00:00:00Z' },
        { index: 1, hash: '0000def456abc789', documentCount: 2, timestamp: '2025-04-28T10:30:00Z' }
    ]
};

// Application state
const appState = {
    wallets: [],
    documents: [],
    templates: [],
    blocks: [],
    selectedDocumentId: null,
    selectedWalletId: null
};

// DOM Elements
const elements = {
    // Tabs
    tabs: document.querySelectorAll('.nav-link'),
    
    // Counters
    documentsCount: document.getElementById('documentsCount'),
    signedDocumentsCount: document.getElementById('signedDocumentsCount'),
    walletsCount: document.getElementById('walletsCount'),
    certWalletsCount: document.getElementById('certWalletsCount'),
    blocksCount: document.getElementById('blocksCount'),
    pendingDocsCount: document.getElementById('pendingDocsCount'),
    
    // Content containers
    recentDocuments: document.getElementById('recentDocuments'),
    recentBlocks: document.getElementById('recentBlocks'),
    walletsTable: document.getElementById('walletsTable'),
    documentsTable: document.getElementById('documentsTable'),
    templatesGrid: document.getElementById('templatesGrid'),
    blocksTable: document.getElementById('blocksTable'),
    
    // Buttons
    createWalletBtn: document.getElementById('createWalletBtn'),
    createFirstWalletBtn: document.getElementById('createFirstWalletBtn'),
    uploadDocumentBtn: document.getElementById('uploadDocumentBtn'),
    uploadFirstDocumentBtn: document.getElementById('uploadFirstDocumentBtn'),
    createTemplateBtn: document.getElementById('createTemplateBtn'),
    createFirstTemplateBtn: document.getElementById('createFirstTemplateBtn'),
    mineBlockBtn: document.getElementById('mineBlockBtn'),
    mineGenesisBlockBtn: document.getElementById('mineGenesisBlockBtn'),
    
    // Modal buttons
    createWalletSubmitBtn: document.getElementById('createWalletSubmitBtn'),
    uploadDocumentSubmitBtn: document.getElementById('uploadDocumentSubmitBtn'),
    signDocumentSubmitBtn: document.getElementById('signDocumentSubmitBtn'),
    createCertificateSubmitBtn: document.getElementById('createCertificateSubmitBtn'),
    
    // Form fields
    walletName: document.getElementById('walletName'),
    documentTitle: document.getElementById('documentTitle'),
    documentFile: document.getElementById('documentFile'),
    signWallet: document.getElementById('signWallet'),
    certificatePassword: document.getElementById('certificatePassword'),
    addToBlockchain: document.getElementById('addToBlockchain'),
    includeTimestamp: document.getElementById('includeTimestamp'),
    certType: document.getElementById('certType'),
    certPassword: document.getElementById('certPassword'),
    certConfirmPassword: document.getElementById('certConfirmPassword'),
    
    // Other
    validationTime: document.getElementById('validationTime'),
    notificationToast: document.getElementById('notificationToast'),
    notificationMessage: document.getElementById('notificationMessage')
};

// Bootstrap modal instances
const modals = {
    createWallet: new bootstrap.Modal(document.getElementById('createWalletModal')),
    uploadDocument: new bootstrap.Modal(document.getElementById('uploadDocumentModal')),
    signDocument: new bootstrap.Modal(document.getElementById('signDocumentModal')),
    createCertificate: new bootstrap.Modal(document.getElementById('createCertificateModal'))
};

// Initialize the application
function initApp() {
    // Load data (in a real app, this would fetch from the API)
    loadMockData();
    
    // Set up event listeners
    setupEventListeners();
    
    // Update UI
    updateUI();
    
    // Set validation time
    elements.validationTime.textContent = new Date().toLocaleString();
}

// Load mock data
function loadMockData() {
    appState.wallets = [...mockData.wallets];
    appState.documents = [...mockData.documents];
    appState.templates = [...mockData.templates];
    appState.blocks = [...mockData.blocks];
}

// Set up event listeners
function setupEventListeners() {
    // Tab switching
    elements.tabs.forEach(tab => {
        tab.addEventListener('click', (e) => {
            e.preventDefault();
            const tabId = e.target.getAttribute('href').substring(1);
            activateTab(tabId);
        });
    });
    
    // Button click handlers
    elements.createWalletBtn.addEventListener('click', () => modals.createWallet.show());
    elements.createFirstWalletBtn.addEventListener('click', () => modals.createWallet.show());
    elements.uploadDocumentBtn.addEventListener('click', () => modals.uploadDocument.show());
    elements.uploadFirstDocumentBtn.addEventListener('click', () => modals.uploadDocument.show());
    elements.mineBlockBtn.addEventListener('click', mineBlock);
    elements.mineGenesisBlockBtn.addEventListener('click', mineBlock);
    
    // Modal form submissions
    elements.createWalletSubmitBtn.addEventListener('click', createWallet);
    elements.uploadDocumentSubmitBtn.addEventListener('click', uploadDocument);
    elements.signDocumentSubmitBtn.addEventListener('click', signDocument);
    elements.createCertificateSubmitBtn.addEventListener('click', createCertificate);
}

// Activate a tab
function activateTab(tabId) {
    elements.tabs.forEach(tab => {
        const target = tab.getAttribute('href').substring(1);
        if (target === tabId) {
            tab.classList.add('active');
            document.getElementById(target).classList.add('show', 'active');
        } else {
            tab.classList.remove('active');
            document.getElementById(target).classList.remove('show', 'active');
        }
    });
}

// Update the UI with current data
function updateUI() {
    updateCounters();
    updateRecentDocuments();
    updateRecentBlocks();
    updateWalletsTable();
    updateDocumentsTable();
    updateTemplatesGrid();
    updateBlocksTable();
}

// Update counter displays
function updateCounters() {
    elements.documentsCount.textContent = appState.documents.length;
    elements.signedDocumentsCount.textContent = `${appState.documents.filter(d => d.status === 'signed').length} signed`;
    
    elements.walletsCount.textContent = appState.wallets.length;
    elements.certWalletsCount.textContent = `${appState.wallets.filter(w => w.hasICPCert).length} with ICP certificates`;
    
    elements.blocksCount.textContent = appState.blocks.length;
    const pendingDocs = appState.documents.filter(doc => doc.status === 'signed' && !doc.blockIndex).length;
    elements.pendingDocsCount.textContent = `${pendingDocs} documents pending`;
}

// Update recent documents list
function updateRecentDocuments() {
    if (appState.documents.length === 0) {
        elements.recentDocuments.innerHTML = '<p class="text-secondary text-center py-3">No documents yet</p>';
        return;
    }
    
    let html = '<ul class="list-unstyled">';
    
    appState.documents.slice(0, 3).forEach(doc => {
        html += `
            <li class="recent-item">
                <div class="d-flex align-items-center">
                    <i class="bi ${doc.status === 'signed' ? 'bi-check-circle text-success' : 'bi-clock text-warning'} recent-item-icon"></i>
                    <span>${doc.title}</span>
                </div>
                <span class="doc-type-badge ${doc.contentType ? `doc-type-${doc.contentType}` : ''}">${doc.contentType}</span>
            </li>
        `;
    });
    
    html += '</ul>';
    elements.recentDocuments.innerHTML = html;
}

// Update recent blocks list
function updateRecentBlocks() {
    if (appState.blocks.length === 0) {
        elements.recentBlocks.innerHTML = '<p class="text-secondary text-center py-3">No blocks yet</p>';
        return;
    }
    
    let html = '<ul class="list-unstyled">';
    
    appState.blocks.slice(0, 3).forEach(block => {
        html += `
            <li class="recent-item">
                <div class="d-flex align-items-center">
                    <i class="bi bi-hdd text-purple recent-item-icon"></i>
                    <span>Block #${block.index}</span>
                </div>
                <span class="text-secondary small">${new Date(block.timestamp).toLocaleDateString()}</span>
            </li>
        `;
    });
    
    html += '</ul>';
    elements.recentBlocks.innerHTML = html;
}

// Update wallets table
function updateWalletsTable() {
    if (appState.wallets.length === 0) {
        elements.walletsTable.innerHTML = `
            <p class="text-secondary text-center py-5">No wallets created yet</p>
            <div class="text-center pb-4">
                <button class="btn btn-primary" id="createFirstWalletBtn">
                    <i class="bi bi-person-plus me-1"></i> Create Your First Wallet
                </button>
            </div>
        `;
        document.getElementById('createFirstWalletBtn').addEventListener('click', () => modals.createWallet.show());
        return;
    }
    
    let html = `
        <div class="table-responsive">
            <table class="table table-hover mb-0">
                <thead class="table-light">
                    <tr>
                        <th>ID</th>
                        <th>Name</th>
                        <th>Certificate</th>
                        <th>Actions</th>
                    </tr>
                </thead>
                <tbody>
    `;
    
    appState.wallets.forEach(wallet => {
        html += `
            <tr>
                <td><span class="wallet-id">${wallet.id}</span></td>
                <td>${wallet.name}</td>
                <td>
                    ${wallet.hasICPCert 
                        ? `<span class="cert-badge cert-badge-valid">${wallet.certType || 'ICP-Brasil'}</span>` 
                        : '<span class="cert-badge cert-badge-none">None</span>'}
                </td>
                <td>
                    ${!wallet.hasICPCert 
                        ? `<button class="btn btn-sm btn-success create-cert-btn" data-wallet-id="${wallet.id}">Create ICP Certificate</button>` 
                        : ''}
                </td>
            </tr>
        `;
    });
    
    html += `
                </tbody>
            </table>
        </div>
    `;
    
    elements.walletsTable.innerHTML = html;
    
    // Add event listeners to the create certificate buttons
    document.querySelectorAll('.create-cert-btn').forEach(btn => {
        btn.addEventListener('click', (e) => {
            appState.selectedWalletId = e.target.getAttribute('data-wallet-id');
            modals.createCertificate.show();
        });
    });
}

// Update documents table
function updateDocumentsTable() {
    if (appState.documents.length === 0) {
        elements.documentsTable.innerHTML = `
            <p class="text-secondary text-center py-5">No documents yet</p>
            <div class="text-center pb-4">
                <button class="btn btn-success" id="uploadFirstDocumentBtn">
                    <i class="bi bi-upload me-1"></i> Upload Your First Document
                </button>
            </div>
        `;
        document.getElementById('uploadFirstDocumentBtn').addEventListener('click', () => modals.uploadDocument.show());
        return;
    }
    
    let html = `
        <div class="table-responsive">
            <table class="table table-hover mb-0">
                <thead class="table-light">
                    <tr>
                        <th>Title</th>
                        <th>Type</th>
                        <th>Status</th>
                        <th>Signed By</th>
                        <th>Actions</th>
                    </tr>
                </thead>
                <tbody>
    `;
    
    appState.documents.forEach(doc => {
        html += `
            <tr>
                <td>${doc.title}</td>
                <td><span class="doc-type-badge ${doc.contentType ? `doc-type-${doc.contentType}` : ''}">${doc.contentType}</span></td>
                <td>
                    ${doc.status === 'signed' 
                        ? '<span class="status-badge status-badge-signed"><i class="bi bi-check-circle me-1"></i> Signed</span>' 
                        : '<span class="status-badge status-badge-pending"><i class="bi bi-clock me-1"></i> Pending</span>'}
                </td>
                <td>${doc.signedBy ? `<span class="wallet-id">${doc.signedBy}</span>` : '-'}</td>
                <td>
                    <div class="btn-group btn-group-sm">
                        ${doc.status !== 'signed' 
                            ? `<button class="btn btn-primary sign-doc-btn" data-doc-id="${doc.id}">Sign</button>` 
                            : `<button class="btn btn-info verify-doc-btn" data-doc-id="${doc.id}">Verify</button>`}
                        <button class="btn btn-secondary">Download</button>
                    </div>
                </td>
            </tr>
        `;
    });
    
    html += `
                </tbody>
            </table>
        </div>
    `;
    
    elements.documentsTable.innerHTML = html;
    
    // Add event listeners to the sign document buttons
    document.querySelectorAll('.sign-doc-btn').forEach(btn => {
        btn.addEventListener('click', (e) => {
            appState.selectedDocumentId = e.target.getAttribute('data-doc-id');
            
            // Populate the wallet dropdown
            const walletSelect = elements.signWallet;
            walletSelect.innerHTML = '<option value="">Select a wallet</option>';
            
            appState.wallets.filter(w => w.hasICPCert).forEach(wallet => {
                const option = document.createElement('option');
                option.value = wallet.id;
                option.textContent = `${wallet.name} (${wallet.certType})`;
                walletSelect.appendChild(option);
            });
            
            modals.signDocument.show();
        });
    });
    
    // Add event listeners to the verify document buttons
    document.querySelectorAll('.verify-doc-btn').forEach(btn => {
        btn.addEventListener('click', (e) => {
            const docId = e.target.getAttribute('data-doc-id');
            verifyDocument(docId);
        });
    });
}

// Update templates grid
function updateTemplatesGrid() {
    if (appState.templates.length === 0) {
        elements.templatesGrid.innerHTML = `
            <div class="col-12 text-center py-5">
                <p class="text-secondary mb-3">No templates available</p>
                <button class="btn btn-primary" id="createFirstTemplateBtn">
                    <i class="bi bi-file-earmark-plus me-1"></i> Create Your First Template
                </button>
            </div>
        `;
        document.getElementById('createFirstTemplateBtn').addEventListener('click', () => {
            showNotification('Template creation is not implemented in this demo', 'info');
        });
        return;
    }
    
    let html = '';
    
    appState.templates.forEach(template => {
        html += `
            <div class="col-md-6 col-lg-4">
                <div class="card border-0 shadow-sm template-card">
                    <div class="card-body template-card-body">
                        <h3 class="h5 fw-semibold text-dark mb-2">${template.name}</h3>
                        <p class="text-secondary mb-3">${template.description}</p>
                        <div class="mb-3">
                            <span class="doc-type-badge ${template.type ? `doc-type-${template.type}` : ''}">${template.type}</span>
                        </div>
                        <div class="template-card-footer d-flex gap-2">
                            <button class="btn btn-primary flex-grow-1">
                                <i class="bi bi-file-earmark-text me-1"></i> Use Template
                            </button>
                            <button class="btn btn-outline-secondary">Preview</button>
                        </div>
                    </div>
                </div>
            </div>
        `;
    });
    
    elements.templatesGrid.innerHTML = html;
}

// Update blocks table
function updateBlocksTable() {
    if (appState.blocks.length === 0) {
        elements.blocksTable.innerHTML = `
            <p class="text-secondary text-center py-5">No blocks in the blockchain yet</p>
            <div class="text-center pb-4">
                <button class="btn btn-purple" id="mineGenesisBlockBtn">
                    <i class="bi bi-hdd me-1"></i> Mine Genesis Block
                </button>
            </div>
        `;
        document.getElementById('mineGenesisBlockBtn').addEventListener('click', mineBlock);
        return;
    }
    
    let html = `
        <div class="table-responsive">
            <table class="table table-hover mb-0">
                <thead class="table-light">
                    <tr>
                        <th>Block #</th>
                        <th>Hash</th>
                        <th>Documents</th>
                        <th>Timestamp</th>
                        <th>Actions</th>
                    </tr>
                </thead>
                <tbody>
    `;
    
    appState.blocks.forEach(block => {
        html += `
            <tr>
                <td>${block.index}</td>
                <td><span class="block-hash" title="${block.hash}">${block.hash}</span></td>
                <td>${block.documentCount}</td>
                <td>${new Date(block.timestamp).toLocaleString()}</td>
                <td>
                    <button class="btn btn-sm btn-primary">View Details</button>
                </td>
            </tr>
        `;
    });
    
    html += `
                </tbody>
            </table>
        </div>
    `;
    
    elements.blocksTable.innerHTML = html;
}

// Create a new wallet
function createWallet() {
    const name = elements.walletName.value.trim();
    
    if (!name) {
        showNotification('Please enter a wallet name', 'error');
        return;
    }
    
    const newWallet = {
        id: generateId(),
        name: name,
        hasICPCert: false
    };
    
    appState.wallets.push(newWallet);
    modals.createWallet.hide();
    elements.walletName.value = '';
    
    updateUI();
    showNotification('Wallet created successfully');
}

// Upload a document
function uploadDocument() {
    const title = elements.documentTitle.value.trim();
    const file = elements.documentFile.files[0];
    
    if (!title) {
        showNotification('Please enter a document title', 'error');
        return;
    }
    
    if (!file) {
        showNotification('Please select a file', 'error');
        return;
    }
    
    // Get file extension
    const fileExtension = file.name.split('.').pop().toLowerCase();
    let contentType;
    
    switch (fileExtension) {
        case 'pdf':
            contentType = 'pdf';
            break;
        case 'html':
        case 'htm':
            contentType = 'html';
            break;
        case 'md':
        case 'markdown':
            contentType = 'md';
            break;
        default:
            contentType = fileExtension;
    }
    
    const newDocument = {
        id: generateId(),
        title: title,
        status: 'pending',
        contentType: contentType
    };
    
    appState.documents.push(newDocument);
    modals.uploadDocument.hide();
    elements.documentTitle.value = '';
    elements.documentFile.value = '';
    
    updateUI();
    showNotification('Document uploaded successfully');
}

// Sign a document
function signDocument() {
    const walletId = elements.signWallet.value;
    const password = elements.certificatePassword.value;
    const addToBlockchain = elements.addToBlockchain.checked;
    const includeTimestamp = elements.includeTimestamp.checked;
    
    if (!walletId) {
        showNotification('Please select a wallet', 'error');
        return;
    }
    
    if (!password) {
        showNotification('Please enter the certificate password', 'error');
        return;
    }
    
    const documentId = appState.selectedDocumentId;
    if (!documentId) {
        showNotification('No document selected', 'error');
        return;
    }
    
    // Update the document
    appState.documents = appState.documents.map(doc => {
        if (doc.id === documentId) {
            return {
                ...doc,
                status: 'signed',
                signedBy: walletId,
                signatureTime: new Date().toISOString(),
                addedToBlockchain: addToBlockchain,
                hasTimestamp: includeTimestamp
            };
        }
        return doc;
    });
    
    modals.signDocument.hide();
    elements.signWallet.value = '';
    elements.certificatePassword.value = '';
    
    updateUI();
    showNotification('Document signed successfully');
}

// Create an ICP certificate
function createCertificate() {
    const certType = elements.certType.value;
    const password = elements.certPassword.value;
    const confirmPassword = elements.certConfirmPassword.value;
    
    if (!password) {
        showNotification('Please enter a certificate password', 'error');
        return;
    }
    
    if (password !== confirmPassword) {
        showNotification('Passwords do not match', 'error');
        return;
    }
    
    const walletId = appState.selectedWalletId;
    if (!walletId) {
        showNotification('No wallet selected', 'error');
        return;
    }
    
    // Update the wallet
    appState.wallets = appState.wallets.map(wallet => {
        if (wallet.id === walletId) {
            return {
                ...wallet,
                hasICPCert: true,
                certType: certType
            };
        }
        return wallet;
    });
    
    modals.createCertificate.hide();
    elements.certPassword.value = '';
    elements.certConfirmPassword.value = '';
    
    updateUI();
    showNotification('ICP-Brasil certificate created successfully');
}

// Mine a block
function mineBlock() {
    const pendingDocs = appState.documents.filter(doc => doc.status === 'signed' && !doc.blockIndex);
    
    if (pendingDocs.length === 0) {
        showNotification('No pending documents to mine', 'error');
        return;
    }
    
    const newBlockIndex = appState.blocks.length;
    const newBlock = {
        index: newBlockIndex,
        hash: '0000' + generateId(),
        documentCount: pendingDocs.length,
        timestamp: new Date().toISOString()
    };
    
    appState.blocks.push(newBlock);
    
    // Update documents to show they're now in a block
    appState.documents = appState.documents.map(doc => {
        if (doc.status === 'signed' && !doc.blockIndex) {
            return {
                ...doc,
                blockIndex: newBlockIndex
            };
        }
        return doc;
    });
    
    updateUI();
    showNotification('Block mined successfully');
}

// Verify a document
function verifyDocument(docId) {
    const doc = appState.documents.find(d => d.id === docId);
    
    if (!doc) {
        showNotification('Document not found', 'error');
        return;
    }
    
    if (doc.status !== 'signed') {
        showNotification('Document is not signed', 'error');
        return;
    }
    
    // Check if document is in blockchain
    const inBlockchain = doc.blockIndex !== undefined;
    
    let message = 'Document is valid and properly signed';
    if (inBlockchain) {
        message += ' and verified in the blockchain';
    }
    
    showNotification(message);
}

// Show a notification
function showNotification(message, type = 'success') {
    elements.notificationMessage.textContent = message;
    elements.notificationToast.classList.remove('toast-success', 'toast-error');
    elements.notificationToast.classList.add(`toast-${type}`);
    
    const toast = new bootstrap.Toast(elements.notificationToast);
    toast.show();
}

// Generate a random ID
function generateId() {
    return Math.random().toString(36).substring(2, 10);
}

// Initialize the application when the DOM is loaded
document.addEventListener('DOMContentLoaded', initApp);
