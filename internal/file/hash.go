package file

import (
	"crypto/sha256"
)

func HashChunk(chunk []byte) []byte {
	hash := sha256.Sum256(chunk)
	return hash[:]
}

func HashChunks(chunks [][]byte) [][]byte {
	var hashes [][]byte
	for _, chunk := range chunks {
		hashes = append(hashes, HashChunk(chunk))
	}
	return hashes
}
