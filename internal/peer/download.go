// Package peer implements the peer-to-peer file sharing functionality.
// It provides both client and server capabilities for sharing files between peers.
package peer

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"

	"github.com/timskillet/go-share/internal/file"
)

type Peer struct {
	Address string `json:"address"`
	Port    int    `json:"port"`
}

// DownloadChunk downloads a specific chunk from a peer
func DownloadChunk(peer Peer, chunkIndex int) ([]byte, error) {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", peer.Address, peer.Port))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to peer: %v", err)
	}
	defer conn.Close()

	// Send chunk request
	request := struct {
		ChunkIndex int `json:"chunkIndex"`
	}{
		ChunkIndex: chunkIndex,
	}

	if err := json.NewEncoder(conn).Encode(request); err != nil {
		return nil, fmt.Errorf("failed to send chunk request: %v", err)
	}

	// Read chunk data
	data, err := io.ReadAll(conn)
	if err != nil {
		return nil, fmt.Errorf("failed to read chunk data: %v", err)
	}

	return data, nil
}

// DownloadFile downloads a file from a peer using its manifest.
// It connects to the specified peer, requests each chunk, and assembles them into the output file.
// The outputPath parameter specifies where the downloaded file should be saved.
func DownloadFile(manifest *file.Manifest, peerAddress string, peerPort int, outputPath string) error {
	// Create output directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	// Create output file
	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %v", err)
	}
	defer outFile.Close()

	// Download each chunk
	for i, chunk := range manifest.Chunks {
		// Connect to peer
		conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", peerAddress, peerPort))
		if err != nil {
			return fmt.Errorf("failed to connect to peer: %v", err)
		}
		defer conn.Close()

		// Send chunk request
		req := struct {
			ChunkIndex int `json:"chunkIndex"`
		}{
			ChunkIndex: i,
		}
		if err := json.NewEncoder(conn).Encode(req); err != nil {
			return fmt.Errorf("failed to send chunk request: %v", err)
		}

		// Read chunk data
		chunkData := make([]byte, chunk.Size)
		if _, err := io.ReadFull(conn, chunkData); err != nil {
			return fmt.Errorf("failed to read chunk data: %v", err)
		}

		// Verify chunk hash
		if !file.VerifyChunk(chunk, chunkData) {
			return fmt.Errorf("chunk hash verification failed")
		}

		// Write chunk to output file
		if _, err := outFile.Write(chunkData); err != nil {
			return fmt.Errorf("failed to write chunk to file: %v", err)
		}
	}

	return nil
}
