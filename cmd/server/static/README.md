# DocChain Frontend

This is the frontend for the DocChain application, a blockchain-based document signing system with ICP-Brasil support.

## Overview

The frontend is built using:
- HTML5
- CSS3 with Bootstrap 5
- JavaScript (vanilla)

It provides a user-friendly interface for interacting with the DocChain backend API, allowing users to:
- Create and manage wallets
- Upload and sign documents
- Verify document signatures
- Create and use ICP-Brasil certificates
- View blockchain information
- Mine blocks

## Structure

- `index.html` - Main HTML file
- `css/styles.css` - Custom CSS styles
- `js/app.js` - Main application logic
- `js/api.js` - API client for communicating with the backend

## Features

### Home Dashboard
- Overview of documents, wallets, and blockchain status
- Recent documents and blocks

### Wallets Management
- Create new wallets
- Create ICP-Brasil certificates for wallets

### Document Management
- Upload documents
- Sign documents with ICP-Brasil certificates
- Verify document signatures
- Download documents

### Templates
- View available document templates
- Create new templates

### Blockchain
- View blockchain blocks
- Mine new blocks
- Validate blockchain integrity

## Integration with Backend

The frontend communicates with the backend API using the API client in `js/api.js`. The API endpoints are:

- `/api/wallet/*` - Wallet management
- `/api/document/*` - Document management
- `/api/blockchain/*` - Blockchain operations
- `/api/icpbrasil/*` - ICP-Brasil specific operations

## Development

To modify the frontend:

1. Edit the HTML, CSS, or JavaScript files as needed
2. The server will automatically serve the updated files

## Deployment

The frontend is served by the Go backend server from the `static` directory. No separate deployment is needed.
