// Package main implements the command-line client for the peer-to-peer file sharing system.
// It provides commands for uploading and downloading files through the network.
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/timskillet/go-share/internal/file"
	"github.com/timskillet/go-share/internal/peer"
)

var (
	chunkSize int64
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "go-share",
	Short: "A peer-to-peer file sharing application",
	Long: `go-share is a command-line tool for sharing files in a peer-to-peer network.
It allows users to upload files to the network and download files from other peers.`,
}

// uploadCmd represents the upload command
var uploadCmd = &cobra.Command{
	Use:   "upload [file]",
	Short: "Upload a file to the network",
	Long: `Upload a file to the peer-to-peer network. The file will be split into chunks
and made available for other peers to download. A manifest file will be created
with the same name as the original file plus a .manifest extension.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		filePath := args[0]

		// Create manifest for the file
		manifest, err := file.CreateManifest(filePath, file.DefaultChunkSize)
		if err != nil {
			fmt.Printf("Error creating manifest: %v\n", err)
			return
		}

		// Save manifest
		if err := file.SaveManifest(manifest, filePath); err != nil {
			fmt.Printf("Error saving manifest: %v\n", err)
			return
		}

		// Start file server in background
		go func() {
			if err := peer.StartFileServer(filePath); err != nil {
				fmt.Printf("Error starting file server: %v\n", err)
				return
			}
		}()

		// Announce file to tracker
		announceReq := struct {
			FileHash string `json:"fileHash"`
			Address  string `json:"address"`
			Port     int    `json:"port"`
		}{
			FileHash: manifest.FileHash,
			Address:  "localhost",
			Port:     9000,
		}

		data, err := json.Marshal(announceReq)
		if err != nil {
			fmt.Printf("Error marshaling announce request: %v\n", err)
			return
		}

		resp, err := http.Post("http://localhost:8080/announce", "application/json", bytes.NewBuffer(data))
		if err != nil {
			fmt.Printf("Error announcing file: %v\n", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			fmt.Printf("Error announcing file: %s\n", resp.Status)
			return
		}

		fmt.Printf("File uploaded successfully. Manifest saved as %s.manifest\n", filePath)
		fmt.Println("Keep this terminal open to serve the file to other peers.")

		// Block to keep the server running
		select {}
	},
}

// downloadCmd represents the download command
var downloadCmd = &cobra.Command{
	Use:   "download [manifest]",
	Short: "Download a file using its manifest",
	Long: `Download a file using its manifest file. The manifest contains information
about the file's chunks and where to find them. The file will be downloaded
from available peers and saved in the same directory as the manifest.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		manifestPath := args[0]

		// Load manifest
		manifest, err := file.LoadManifest(manifestPath)
		if err != nil {
			return fmt.Errorf("error loading manifest: %v", err)
		}

		// Get list of peers from tracker
		resp, err := http.Get(fmt.Sprintf("http://localhost:8080/peers?fileHash=%s", manifest.FileHash))
		if err != nil {
			return fmt.Errorf("error getting peers: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("error getting peers: %s", resp.Status)
		}

		var peersResp struct {
			Peers []struct {
				Address string `json:"address"`
				Port    int    `json:"port"`
			} `json:"peers"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&peersResp); err != nil {
			return fmt.Errorf("error decoding peers response: %v", err)
		}

		if len(peersResp.Peers) == 0 {
			return fmt.Errorf("no peers found for this file")
		}

		// Download file
		downloadsDir := "downloads"
		if err := os.MkdirAll(downloadsDir, 0755); err != nil {
			return fmt.Errorf("error creating downloads directory: %v", err)
		}
		outputPath := filepath.Join(downloadsDir, manifest.FileName)
		if err := peer.DownloadFile(manifest, peersResp.Peers[0].Address, peersResp.Peers[0].Port, outputPath); err != nil {
			return fmt.Errorf("error downloading file: %v", err)
		}

		fmt.Printf("File downloaded successfully to %s\n", outputPath)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(uploadCmd)
	rootCmd.AddCommand(downloadCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
