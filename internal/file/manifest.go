// Package file implements file handling functionality for the peer-to-peer file sharing system.
// It provides utilities for creating file manifests, handling chunks, and managing file operations.
package file

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
)

// Chunk represents a portion of a file that can be shared independently.
// Each chunk has a unique hash and a specific size within the file.
type Chunk struct {
	Hash string `json:"hash"` // SHA-256 hash of the chunk data
	Size int64  `json:"size"` // Size of the chunk in bytes
}

// Manifest represents the metadata for a shared file.
// It contains information about the file and its chunks.
type Manifest struct {
	FileName  string  `json:"fileName"`  // Original name of the file
	FileSize  int64   `json:"fileSize"`  // Total size of the file in bytes
	ChunkSize int64   `json:"chunkSize"` // Size of each chunk in bytes
	Chunks    []Chunk `json:"chunks"`    // List of chunks that make up the file
	FileHash  string  `json:"fileHash"`  // SHA-256 hash of the entire file
}

// DefaultChunkSize is the default size for file chunks (1MB).
const DefaultChunkSize = 1024 * 1024

// CreateManifest creates a new manifest for a file.
// It splits the file into chunks and calculates their hashes.
// The chunkSize parameter determines how large each chunk should be.
func CreateManifest(filePath string, chunkSize int64) (*Manifest, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return nil, err
	}

	manifest := &Manifest{
		FileName:  fileInfo.Name(),
		FileSize:  fileInfo.Size(),
		ChunkSize: chunkSize,
	}

	// Calculate file hash
	fileHash := sha256.New()
	if _, err := file.Seek(0, 0); err != nil {
		return nil, err
	}
	if _, err := file.WriteTo(fileHash); err != nil {
		return nil, err
	}
	manifest.FileHash = fmt.Sprintf("%x", fileHash.Sum(nil))

	// Create chunks
	numChunks := (fileInfo.Size() + chunkSize - 1) / chunkSize
	manifest.Chunks = make([]Chunk, numChunks)

	for i := int64(0); i < numChunks; i++ {
		chunkSize := chunkSize
		if i == numChunks-1 {
			chunkSize = fileInfo.Size() - (i * chunkSize)
		}

		chunk := Chunk{
			Size: chunkSize,
		}

		// Calculate chunk hash
		chunkHash := sha256.New()
		if _, err := file.Seek(i*chunkSize, 0); err != nil {
			return nil, err
		}
		if _, err := file.WriteTo(chunkHash); err != nil {
			return nil, err
		}
		chunk.Hash = fmt.Sprintf("%x", chunkHash.Sum(nil))

		manifest.Chunks[i] = chunk
	}

	return manifest, nil
}

// SaveManifest saves a manifest to a file.
// The manifest is saved in JSON format with the same name as the original file
// plus a .manifest extension.
func SaveManifest(manifest *Manifest, filePath string) error {
	manifestPath := filePath + ".manifest"
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(manifestPath, data, 0644)
}

// LoadManifest loads a manifest from a file.
// It reads and parses the JSON data into a Manifest struct.
func LoadManifest(manifestPath string) (*Manifest, error) {
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, err
	}

	var manifest Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, err
	}

	return &manifest, nil
}
