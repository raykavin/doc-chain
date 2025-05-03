# Doc Chain

A blockchain-based document signing and verification system that allows users to securely sign documents, verify signatures, and maintain an immutable record of signed documents.

## Features

- Blockchain-based document storage and verification
- Digital signatures using ECDSA (Elliptic Curve Digital Signature Algorithm)
- RESTful API for easy integration
- Proof-of-work consensus mechanism
- Merkle tree for efficient document verification

## Architecture

The application follows clean architecture and SOLID principles with a clear separation of concerns:

- **blockchain package**: Core blockchain functionality (blocks, documents, etc.)
- **crypto package**: Cryptographic utilities (hashing, signatures, wallets)
- **server package**: HTTP server and API endpoints
- **models package**: Shared data structures and models
- **utils package**: Common utility functions

## API Endpoints

### Wallet Endpoints

- `POST /api/wallet/new`: Create a new wallet
- `GET /api/wallet/{id}`: Get wallet information

### Document Endpoints

- `POST /api/document/sign`: Sign a document and add it to the blockchain
- `GET /api/document/{id}`: Get document information
- `GET /api/document/verify/{id}`: Verify a document signature

### Blockchain Endpoints

- `GET /api/blockchain`: Get blockchain information
- `POST /api/blockchain/mine`: Mine a new block
- `GET /api/blockchain/blocks`: Get all blocks
- `GET /api/blockchain/block/{hash}`: Get a block by hash
- `GET /api/blockchain/pending`: Get pending documents
- `GET /api/blockchain/validate`: Validate the blockchain

## Getting Started

### Prerequisites

- Go 1.16 or later
- [Gorilla Mux](https://github.com/gorilla/mux) for routing
- [Cobra](https://github.com/spf13/cobra) for CLI
- [Viper](https://github.com/spf13/viper) for configuration

### Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/raykavin/doc-chain.git
   cd doc-chain/
   ```

2. Install dependencies:
   ```bash
   go mod tidy
   ```

3. Build the server:
   ```bash
   go build -o blockchain-server ./cmd/server
   ```

4. Build the CLI:
   ```bash
   go build -o blockchain-cli ./cmd/cli
   ```

5. Run the server:
   ```bash
   ./blockchain-server
   ```
   
   By default, the server will run on port 8080 with a mining difficulty of 4. You can change these settings using command-line flags:
   ```bash
   ./blockchain-server --port=9000 --difficulty=5
   ```

### Using the CLI

The blockchain-cli tool provides a convenient way to interact with the blockchain server:

```bash
# Get general help
./blockchain-cli --help

# Create a new wallet
./blockchain-cli wallet create --output wallet.json

# Sign a document
./blockchain-cli document sign --content "This is a test document" --key-file wallet.json

# Mine a new block
./blockchain-cli blockchain mine

# Validate the blockchain
./blockchain-cli blockchain validate
```

For more information on using the CLI, see the [CLI documentation](cmd/cli/README.md).

## Usage Examples

### Using the REST API

#### Creating a Wallet

```bash
curl -X POST http://localhost:8080/api/wallet/new
```

Response:
```json
{
  "success": true,
  "message": "Wallet created successfully",
  "data": {
    "id": "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
    "publicKey": "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
    "privateKey": "-----BEGIN EC PRIVATE KEY-----\n..."
  }
}
```

#### Signing a Document

```bash
curl -X POST -H "Content-Type: application/json" -d '{
  "content": "This is a test document",
  "key": "-----BEGIN EC PRIVATE KEY-----\n..."
}' http://localhost:8080/api/document/sign
```

Response:
```json
{
  "success": true,
  "message": "Document signed and added to pending documents",
  "data": {
    "documentId": "a948904f2f0f479b8f8197694b30184b0d2ed1c1cd2a1ec0fb85d299a192a447",
    "signedBy": "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
    "timestamp": "1620000000"
  }
}
```

#### Mining a Block

```bash
curl -X POST http://localhost:8080/api/blockchain/mine
```

Response:
```json
{
  "success": true,
  "message": "Block mined successfully",
  "data": {
    "index": 1,
    "hash": "000012345abc...",
    "documents": 1,
    "timestamp": 1620000000,
    "merkleRoot": "a948904f2f0f479b8f8197694b30184b0d2ed1c1cd2a1ec0fb85d299a192a447",
    "nonce": 12345
  }
}
```

### Using the CLI

#### Complete Workflow Example

```bash
# Create a new wallet and save it to a file
./blockchain-cli wallet create --output wallet.json

# Sign a document using the wallet
./blockchain-cli document sign --content "This is a test document" --key-file wallet.json
# Document ID: a948904f2f0f479b8f8197694b30184b0d2ed1c1cd2a1ec0fb85d299a192a447

# View pending documents
./blockchain-cli blockchain pending

# Mine a new block to include the document in the blockchain
./blockchain-cli blockchain mine

# Verify the document
./blockchain-cli document verify a948904f2f0f479b8f8197694b30184b0d2ed1c1cd2a1ec0fb85d299a192a447

# View all blocks in the blockchain
./blockchain-cli blockchain blocks --verbose

# Validate the integrity of the blockchain
./blockchain-cli blockchain validate
```


## 🤝 Contributing

Contributions to DocChain are welcome! Here are some ways you can contribute:

1. Report bugs and suggest features by opening issues
2. Submit pull requests with bug fixes or new features
3. Improve documentation
4. Share your custom strategies with the community

## 📄License

MIT License © [Raykavin Meireles](https://github.com/raykavin)

DocChain is licensed under the MIT License. See the [LICENSE](LICENSE.md) file for details.

---
## 📬 Contact

Feel free to reach out for support or collaboration:  
**Email**: [raykavin.meireles@gmail.com](mailto:raykavin.meireles@gmail.com)  
**GitHub**: [@raykavin](https://github.com/raykavin)\
**LinkedIn**: [@raykavin.dev](https://www.linkedin.com/in/raykavin-dev)\
**Instagram**: [@raykavin.dev](https://www.instagram.com/raykavin.dev)
