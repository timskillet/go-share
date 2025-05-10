# Go-Share: Peer-to-Peer File Sharing System

Go-Share is a decentralized peer-to-peer file sharing system written in Go. It allows users to share files directly between peers without relying on a central server for file storage. The system uses a tracker server only for peer discovery, while the actual file transfer happens directly between peers.

## Architecture

The system consists of three main components:

### 1. Tracker Server
- Acts as a central registry for peer discovery
- Maintains a database of which peers have which files
- Provides HTTP endpoints for peers to:
  - Announce when they have a file to share
  - Query which peers have a specific file
- Runs on a configurable port (default: 8080)

### 2. Peer Server
- Runs on each peer that wants to share files
- Listens for incoming chunk requests on port 9000
- Handles file chunk requests from other peers
- Serves file chunks over TCP connections
- Verifies chunk integrity using SHA-256 hashes

### 3. File Management System
- Handles file chunking and manifest creation
- Splits files into manageable chunks (default: 1MB)
- Creates and manages file manifests containing:
  - File metadata (name, size)
  - Chunk information (hashes, sizes)
  - File integrity verification
- Provides utilities for chunk verification and integrity checking

## How It Works

### Uploading a File
1. The file is split into chunks of configurable size
2. A manifest file is created containing metadata and chunk information
3. The peer announces to the tracker that it has the file
4. The peer server starts listening for chunk requests

### Downloading a File
1. The peer requests the file manifest from the tracker
2. The tracker returns a list of peers that have the file
3. The peer connects to other peers to download chunks
4. Chunks are verified using their SHA-256 hashes
5. The file is reassembled from the downloaded chunks

## Security Features
- SHA-256 hashing for file and chunk integrity verification
- Chunk-level verification to ensure data integrity
- Direct peer-to-peer connections for file transfer
- No central storage of file contents

## Usage

### Starting the Tracker Server
```bash
go run cmd/tracker/main.go
```

### Sharing a File
```bash
go run cmd/peer/main.go share <file_path>
```

### Downloading a File
```bash
go run cmd/peer/main.go download <manifest_path>
```

## Project Structure
```
.
├── cmd/
│   ├── tracker/    # Tracker server implementation
│   └── peer/       # Peer client implementation
├── internal/
│   ├── tracker/    # Tracker server logic
│   ├── peer/       # Peer server and client logic
│   └── file/       # File handling and chunking
├── downloads/      # Default download directory
└── main.go        # Main entry point
```

## Dependencies
- Go 1.16 or higher
- Standard library packages:
  - `net/http` for tracker server
  - `net` for peer-to-peer communication
  - `crypto/sha256` for file integrity

## License
This project is licensed under the MIT License - see the LICENSE file for details.