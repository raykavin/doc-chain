# DocChain

A blockchain-based document signing system with ICP-Brasil compliance.

## Overview

DocChain is a secure document signing platform that combines blockchain technology with ICP-Brasil digital certificates to provide legally valid document signatures with immutable proof of existence and integrity.

## Features

- **Document Signing**: Sign documents using ICP-Brasil compliant digital certificates
- **Blockchain Integration**: Store document hashes and signature information on a blockchain for immutability
- **Certificate Management**: Create and manage digital certificates
- **Signature Verification**: Verify document signatures and their blockchain records
- **Revocation Checking**: Check certificate revocation status via CRL and OCSP
- **Timestamp Authority**: Optional timestamping via TSA

## Architecture

The application consists of:

- **Backend**: Go-based API server with blockchain and cryptography modules
- **Frontend**: HTML/CSS/JavaScript web interface
- **CLI**: Command-line interface for all operations

## Technical Stack

- **Backend**: Go (Golang)
- **Frontend**: HTML5, CSS3, JavaScript, Bootstrap 5
- **Blockchain**: Custom implementation with Proof of Work
- **Cryptography**: CMS/PKCS#7 for signatures, SHA-256 for hashing
- **API**: RESTful JSON API

## Getting Started

### Prerequisites

- Go 1.18 or higher
- Git

### Installation

1. Clone the repository:
   ```
   git clone https://github.com/raykavin/docchain.git
   cd docchain
   ```

2. Build the application:
   ```
   go build -o docchain ./cmd/server
   ```

3. Run the server:
   ```
   ./docchain
   ```

4. Access the web interface at http://localhost:8080

### Using the CLI

The CLI provides access to all functionality:

```
# Create a new wallet
./docchain wallet create --name "My Wallet"

# Sign a document
./docchain document sign --file document.pdf --wallet my-wallet-id

# Mine a new block
./docchain blockchain mine
```

## ICP-Brasil Compliance

DocChain implements the following ICP-Brasil requirements:

- A1 certificate support (.pfx/.p12)
- CMS/PKCS#7 detached signatures
- Complete certificate chain embedding
- Timestamp Authority integration
- CRL and OCSP revocation checking

## License

This project is licensed under the MIT License - see the LICENSE.md file for details.
