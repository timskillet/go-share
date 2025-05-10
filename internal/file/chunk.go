// Package file implements file handling functionality for the peer-to-peer file sharing system.
// It provides utilities for creating file manifests, handling chunks, and managing file operations.
package file

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
)

// GetChunk retrieves a specific chunk from a file.
// It reads the chunk data from the file and returns it as a byte slice.
// The chunk is identified by its index in the manifest's chunks array.
func GetChunk(filePath string, chunk Chunk) ([]byte, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Calculate the offset for this chunk
	offset := int64(0)
	for i := 0; i < len(chunk.Hash); i++ {
		offset += chunk.Size
	}

	// Seek to the chunk's position
	if _, err := file.Seek(offset, 0); err != nil {
		return nil, err
	}

	// Read the chunk data
	data := make([]byte, chunk.Size)
	if _, err := io.ReadFull(file, data); err != nil {
		return nil, err
	}

	// Verify the chunk hash
	hash := sha256.Sum256(data)
	if fmt.Sprintf("%x", hash) != chunk.Hash {
		return nil, fmt.Errorf("chunk hash verification failed")
	}

	return data, nil
}

// WriteChunk writes a chunk of data to a file at the specified offset.
// It verifies the chunk's hash before writing to ensure data integrity.
func WriteChunk(file *os.File, chunk Chunk, data []byte) error {
	// Verify the chunk hash
	hash := sha256.Sum256(data)
	if fmt.Sprintf("%x", hash) != chunk.Hash {
		return fmt.Errorf("chunk hash verification failed")
	}

	// Write the chunk data
	if _, err := file.Write(data); err != nil {
		return err
	}

	return nil
}

// VerifyChunk verifies that a chunk's data matches its hash.
// It returns true if the hash matches, false otherwise.
func VerifyChunk(chunk Chunk, data []byte) bool {
	hash := sha256.Sum256(data)
	return fmt.Sprintf("%x", hash) == chunk.Hash
}
