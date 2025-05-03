/**
 * DocChain - Blockchain Document Signing with ICP-Brasil Support
 * Main application JavaScript
 */

// API client instance
const apiClient = api;

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
async function initApp() {
    try {
        // Load data from API
        await loadData();
        
        // Set up event listeners
        setupEventListeners();
        
        // Update UI
        updateUI();
        
        // Set validation time
        elements.validationTime.textContent = new Date().toLocaleString();
    } catch (error) {
        console.error('Falha ao inicializar aplicativo:', error);
        showNotification('Falha ao carregar dados do servidor', 'error');
    }
}

// Load data from API
async function loadData() {
    try {
        // Load wallets
        const walletsResponse = await apiClient.request('/api/wallet/list');
        appState.wallets = walletsResponse.wallets || [];
        
        // Load documents
        const documentsResponse = await apiClient.request('/api/document/list');
        appState.documents = documentsResponse.documents || [];
        
        // Load templates
        const templatesResponse = await apiClient.request('/api/template/list');
        appState.templates = templatesResponse.templates || [];
        
        // Load blocks
        const blocksResponse = await apiClient.request('/api/blockchain/blocks');
        appState.blocks = blocksResponse.blocks || [];
    } catch (error) {
        console.error('Erro ao carregar dados:', error);
        // If API fails, use empty arrays
        appState.wallets = [];
        appState.documents = [];
        appState.templates = [];
        appState.blocks = [];
    }
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
    elements.signedDocumentsCount.textContent = `${appState.documents.filter(d => d.status === 'signed').length} assinados`;
    
    elements.walletsCount.textContent = appState.wallets.length;
    elements.certWalletsCount.textContent = `${appState.wallets.filter(w => w.hasCert).length} com certificados ICP`;
    
    elements.blocksCount.textContent = appState.blocks.length;
    const pendingDocs = appState.documents.filter(doc => doc.status === 'signed' && !doc.blockIndex).length;
    elements.pendingDocsCount.textContent = `${pendingDocs} documentos pendentes`;
}

// Update recent documents list
function updateRecentDocuments() {
    if (appState.documents.length === 0) {
        elements.recentDocuments.innerHTML = '<p class="text-secondary text-center py-3">Nenhum documento ainda</p>';
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
        elements.recentBlocks.innerHTML = '<p class="text-secondary text-center py-3">Nenhum bloco ainda</p>';
        return;
    }
    
    let html = '<ul class="list-unstyled">';
    
    appState.blocks.slice(0, 3).forEach(block => {
        html += `
            <li class="recent-item">
                <div class="d-flex align-items-center">
                    <i class="bi bi-hdd text-purple recent-item-icon"></i>
                    <span>Bloco #${block.index}</span>
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
            <p class="text-secondary text-center py-5">Nenhuma carteira criada ainda</p>
            <div class="text-center pb-4">
                <button class="btn btn-primary" id="createFirstWalletBtn">
                    <i class="bi bi-person-plus me-1"></i> Criar Sua Primeira Carteira
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
                        <th>Nome</th>
                        <th>Certificado</th>
                        <th>Ações</th>
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
                    ${wallet.hasCert 
                        ? `<span class="cert-badge cert-badge-valid">${wallet.certType || 'ICP-Brasil'}</span>` 
                        : '<span class="cert-badge cert-badge-none">Nenhum</span>'}
                </td>
                <td>
                    ${!wallet.hasCert 
                        ? `<button class="btn btn-sm btn-success create-cert-btn" data-wallet-id="${wallet.id}">Criar Certificado ICP</button>` 
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
            <p class="text-secondary text-center py-5">Nenhum documento ainda</p>
            <div class="text-center pb-4">
                <button class="btn btn-success" id="uploadFirstDocumentBtn">
                    <i class="bi bi-upload me-1"></i> Envie Seu Primeiro Documento
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
                        <th>Título</th>
                        <th>Tipo</th>
                        <th>Status</th>
                        <th>Assinado Por</th>
                        <th>Ações</th>
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
                        ? '<span class="status-badge status-badge-signed"><i class="bi bi-check-circle me-1"></i> Assinado</span>' 
                        : '<span class="status-badge status-badge-pending"><i class="bi bi-clock me-1"></i> Pendente</span>'}
                </td>
                <td>${doc.signedBy ? `<span class="wallet-id">${doc.signedBy}</span>` : '-'}</td>
                <td>
                    <div class="btn-group btn-group-sm">
                        ${doc.status !== 'signed' 
                            ? `<button class="btn btn-primary sign-doc-btn" data-doc-id="${doc.id}">Assinar</button>` 
                            : `<button class="btn btn-info verify-doc-btn" data-doc-id="${doc.id}">Verificar</button>`}
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
            walletSelect.innerHTML = '<option value="">Selecione uma carteira</option>';
            
            appState.wallets.filter(w => w.hasCert).forEach(wallet => {
                const option = document.createElement('option');
                option.value = wallet.id;
                option.textContent = `${wallet.name} (${wallet.certType || 'ICP-Brasil'})`;
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
                <p class="text-secondary mb-3">Nenhum modelo disponível</p>
                <button class="btn btn-primary" id="createFirstTemplateBtn">
                    <i class="bi bi-file-earmark-plus me-1"></i> Crie Seu Primeiro Modelo
                </button>
            </div>
        `;
        document.getElementById('createFirstTemplateBtn').addEventListener('click', () => {
            showNotification('Indisponível', 'error');
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
                            <span class="doc-type-badge ${template.contentType ? `doc-type-${template.contentType}` : ''}">${template.contentType}</span>
                        </div>
                        <div class="template-card-footer d-flex gap-2">
                            <button class="btn btn-primary flex-grow-1">
                                <i class="bi bi-file-earmark-text me-1"></i> Usar Modelo
                            </button>
                            <button class="btn btn-outline-secondary">Visualizar</button>
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
            <p class="text-secondary text-center py-5">Nenhum bloco na blockchain ainda</p>
            <div class="text-center pb-4">
                <button class="btn btn-purple" id="mineGenesisBlockBtn">
                    <i class="bi bi-hdd me-1"></i> Minerar Bloco Gênesis
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
                        <th>Bloco #</th>
                        <th>Hash</th>
                        <th>Documentos</th>
                        <th>Timestamp</th>
                        <th>Ações</th>
                    </tr>
                </thead>
                <tbody>
    `;
    
    appState.blocks.forEach(block => {
        const documentCount = block.documents ? block.documents.length : 0;
        
        html += `
            <tr>
                <td>${block.index}</td>
                <td><span class="block-hash" title="${block.hash}">${block.hash}</span></td>
                <td>${documentCount}</td>
                <td>${new Date(block.timestamp).toLocaleString()}</td>
                <td>
                    <button class="btn btn-sm btn-primary">Ver Detalhes</button>
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
async function createWallet() {
    const name = elements.walletName.value.trim();
    
    if (!name) {
        showNotification('Por favor, insira um nome para a carteira', 'error');
        return;
    }
    
    try {
        const response = await apiClient.createWallet(name);
        
        if (response && response.wallet) {
            appState.wallets.push(response.wallet);
            modals.createWallet.hide();
            elements.walletName.value = '';
            
            updateUI();
            showNotification('Carteira criada com sucesso');
        } else {
            throw new Error('Resposta inválida do servidor');
        }
    } catch (error) {
        console.error('Falha ao criar carteira:', error);
        showNotification('Falha ao criar carteira: ' + (error.message || 'Erro desconhecido'), 'error');
    }
}

// Upload a document
async function uploadDocument() {
    const title = elements.documentTitle.value.trim();
    const file = elements.documentFile.files[0];
    
    if (!title) {
        showNotification('Por favor, insira um título para o documento', 'error');
        return;
    }
    
    if (!file) {
        showNotification('Por favor, selecione um arquivo', 'error');
        return;
    }
    
    try {
        // Create form data
        const formData = new FormData();
        formData.append('title', title);
        formData.append('file', file);
        
        // Custom request for file upload
        const response = await fetch('/api/document/upload', {
            method: 'POST',
            body: formData
        });
        
        if (!response.ok) {
            const errorData = await response.json();
            throw new Error(errorData.message || 'Falha ao enviar documento');
        }
        
        const data = await response.json();
        
        if (data && data.document) {
            appState.documents.push(data.document);
            modals.uploadDocument.hide();
            elements.documentTitle.value = '';
            elements.documentFile.value = '';
            
            updateUI();
            showNotification('Documento enviado com sucesso');
        } else {
            throw new Error('Resposta inválida do servidor');
        }
    } catch (error) {
        console.error('Falha ao enviar documento:', error);
        showNotification('Falha ao enviar documento: ' + (error.message || 'Erro desconhecido'), 'error');
    }
}

// Sign a document
async function signDocument() {
    const walletId = elements.signWallet.value;
    const password = elements.certificatePassword.value;
    const addToBlockchain = elements.addToBlockchain.checked;
    const includeTimestamp = elements.includeTimestamp.checked;
    
    if (!walletId) {
        showNotification('Por favor, selecione uma carteira', 'error');
        return;
    }
    
    if (!password) {
        showNotification('Por favor, insira a senha do certificado', 'error');
        return;
    }
    
    const documentId = appState.selectedDocumentId;
    if (!documentId) {
        showNotification('Nenhum documento selecionado', 'error');
        return;
    }
    
    try {
        const response = await apiClient.signDocument({
            documentId,
            walletId,
            password,
            addToBlockchain,
            includeTimestamp
        });
        
        if (response && response.success) {
            // Update the document in our state
            const updatedDoc = response.document;
            appState.documents = appState.documents.map(doc => 
                doc.id === documentId ? updatedDoc : doc
            );
            
            modals.signDocument.hide();
            elements.signWallet.value = '';
            elements.certificatePassword.value = '';
            
            updateUI();
            showNotification('Documento assinado com sucesso');
        } else {
            throw new Error(response.message || 'Falha ao assinar documento');
        }
    } catch (error) {
        console.error('Falha ao assinar documento:', error);
        showNotification('Falha ao assinar documento: ' + (error.message || 'Erro desconhecido'), 'error');
    }
}

// Create an ICP certificate
async function createCertificate() {
    const certType = elements.certType.value;
    const password = elements.certPassword.value;
    const confirmPassword = elements.certConfirmPassword.value;
    
    if (!password) {
        showNotification('Por favor, insira uma senha para o certificado', 'error');
        return;
    }
    
    if (password !== confirmPassword) {
        showNotification('As senhas não coincidem', 'error');
        return;
    }
    
    const walletId = appState.selectedWalletId;
    if (!walletId) {
        showNotification('Nenhuma carteira selecionada', 'error');
        return;
    }
    
    try {
        const response = await apiClient.request('/api/icpbrasil/create-certificate', 'POST', {
            walletId,
            certType,
            password
        });
        
        if (response && response.success) {
            // Update the wallet in our state
            appState.wallets = appState.wallets.map(wallet => {
                if (wallet.id === walletId) {
                    return {
                        ...wallet,
                        hasCert: true,
                        certType: certType
                    };
                }
                return wallet;
            });
            
            modals.createCertificate.hide();
            elements.certPassword.value = '';
            elements.certConfirmPassword.value = '';
            
            updateUI();
            showNotification('Certificado ICP-Brasil criado com sucesso');
        } else {
            throw new Error(response.message || 'Falha ao criar certificado');
        }
    } catch (error) {
        console.error('Falha ao criar certificado:', error);
        showNotification('Falha ao criar certificado: ' + (error.message || 'Erro desconhecido'), 'error');
    }
}

// Mine a block
async function mineBlock() {
    try {
        const response = await apiClient.mineBlock();
        
        if (response && response.block) {
            // Add the new block to our state
            appState.blocks.push(response.block);
            
            // Update documents that are now in the block
            if (response.updatedDocuments) {
                response.updatedDocuments.forEach(updatedDoc => {
                    appState.documents = appState.documents.map(doc => 
                        doc.id === updatedDoc.id ? updatedDoc : doc
                    );
                });
            }
            
            updateUI();
            showNotification('Bloco minerado com sucesso');
        } else {
            throw new Error(response.message || 'Falha ao minerar bloco');
        }
    } catch (error) {
        console.error('Falha ao minerar bloco:', error);
        showNotification('Falha ao minerar bloco: ' + (error.message || 'Erro desconhecido'), 'error');
    }
}

// Verify a document
async function verifyDocument(docId) {
    try {
        const response = await apiClient.verifyDocument(docId);
        
        if (response && response.valid) {
            let message = 'Documento é válido e assinado corretamente';
            if (response.inBlockchain) {
                message += ' e verificado na blockchain';
            }
            
            showNotification(message);
        } else {
            showNotification(response.message || 'Documento inválido', 'error');
        }
    } catch (error) {
        console.error('Falha ao verificar documento:', error);
        showNotification('Falha ao verificar documento: ' + (error.message || 'Erro desconhecido'), 'error');
    }
}

// Show a notification
function showNotification(message, type = 'success') {
    elements.notificationMessage.textContent = message;
    elements.notificationToast.classList.remove('toast-success', 'toast-error', 'toast-info');
    elements.notificationToast.classList.add(`toast-${type}`);
    
    const toast = new bootstrap.Toast(elements.notificationToast);
    toast.show();
}

// Initialize the application when the DOM is loaded
document.addEventListener('DOMContentLoaded', initApp);
