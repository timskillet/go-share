// Package peer implements the peer-to-peer file sharing functionality.
// It provides both client and server capabilities for sharing files between peers.
package peer

import (
	"encoding/json"
	"fmt"
	"net"

	"github.com/timskillet/go-share/internal/file"
)

// StartFileServer starts a TCP server that listens for incoming chunk requests.
// It accepts connections on port 9000 and handles them in separate goroutines.
// The server will continue running until an error occurs or the process is terminated.
func StartFileServer(filePath string) error {
	ln, err := net.Listen("tcp", ":9000")
	if err != nil {
		return err
	}
	defer ln.Close()

	fmt.Printf("Peer server started, serving file: %s\n", filePath)
	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}
		go handleConnection(conn, filePath)
	}
}

// ChunkRequest represents a request from a peer to download a specific chunk of a file.
// The ChunkIndex field specifies which chunk of the file is being requested.
type ChunkRequest struct {
	ChunkIndex int `json:"chunkIndex"` // Index of the chunk being requested
}

// handleConnection processes an incoming connection from a peer requesting a file chunk.
// It reads the chunk request, validates it, and sends the requested chunk data.
// The connection is automatically closed when the function returns.
func handleConnection(conn net.Conn, filePath string) {
	defer conn.Close()

	// Read and decode the chunk request
	var req ChunkRequest
	if err := json.NewDecoder(conn).Decode(&req); err != nil {
		fmt.Printf("Error reading chunk request: %v\n", err)
		return
	}

	// Create manifest to get chunk information
	manifest, err := file.CreateManifest(filePath, file.DefaultChunkSize)
	if err != nil {
		fmt.Printf("Error creating manifest: %v\n", err)
		return
	}

	// Find the requested chunk
	if req.ChunkIndex < 0 || req.ChunkIndex >= len(manifest.Chunks) {
		fmt.Printf("Invalid chunk index: %d\n", req.ChunkIndex)
		return
	}

	// Read the chunk data
	chunkData, err := file.GetChunk(filePath, manifest, req.ChunkIndex)
	if err != nil {
		fmt.Printf("Error reading chunk: %v\n", err)
		return
	}

	// Send the chunk data
	if _, err := conn.Write(chunkData); err != nil {
		fmt.Printf("Error sending chunk: %v\n", err)
		return
	}
}
