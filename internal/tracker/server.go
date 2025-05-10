// Package tracker implements the central server that keeps track of which peers have which files.
// It maintains a registry of peers and their shared files, allowing other peers to discover
// where they can download specific files from.
package tracker

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

// Peer represents a node in the network that can serve files.
// It contains the network address and port where the peer can be reached.
type Peer struct {
	Address string `json:"address"` // IP address or hostname of the peer
	Port    int    `json:"port"`    // Port number where the peer is listening
}

// Tracker is the central server that maintains the peer registry.
// It uses a thread-safe map to store which peers have which files.
type Tracker struct {
	mu    sync.RWMutex      // Mutex to protect concurrent access to the peers map
	peers map[string][]Peer // Map of file hashes to list of peers that have the file
}

// NewTracker creates and returns a new Tracker instance with an initialized peers map.
func NewTracker() *Tracker {
	return &Tracker{
		peers: make(map[string][]Peer),
	}
}

// AnnounceRequest represents the data sent by peers when they announce they have a file.
type AnnounceRequest struct {
	FileHash string `json:"fileHash"` // Hash of the file being announced
	Address  string `json:"address"`  // IP address of the announcing peer
	Port     int    `json:"port"`     // Port where the peer is serving the file
}

// PeersResponse represents the data sent back to peers requesting information about a file.
type PeersResponse struct {
	Peers []Peer `json:"peers"` // List of peers that have the requested file
}

// Announce handles HTTP POST requests from peers announcing they have a file.
// It adds the peer to the list of peers that have the specified file.
func (t *Tracker) Announce(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req AnnounceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	peer := Peer{
		Address: req.Address,
		Port:    req.Port,
	}

	// Add peer to the list if not already present
	peers := t.peers[req.FileHash]
	for _, p := range peers {
		if p.Address == peer.Address && p.Port == peer.Port {
			return
		}
	}
	t.peers[req.FileHash] = append(peers, peer)

	w.WriteHeader(http.StatusOK)
}

// GetPeers handles HTTP GET requests from peers looking for other peers that have a file.
// It returns a list of peers that have the requested file.
func (t *Tracker) GetPeers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	fileHash := r.URL.Query().Get("fileHash")
	if fileHash == "" {
		http.Error(w, "Missing fileHash parameter", http.StatusBadRequest)
		return
	}

	t.mu.RLock()
	peers := t.peers[fileHash]
	t.mu.RUnlock()

	response := PeersResponse{
		Peers: peers,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// StartTrackerServer starts the HTTP server that handles peer announcements and queries.
// It listens on the specified port and sets up the necessary HTTP handlers.
func StartTrackerServer(port int) error {
	tracker := NewTracker()
	http.HandleFunc("/announce", tracker.Announce)
	http.HandleFunc("/peers", tracker.GetPeers)
	fmt.Printf("Tracker listening on port %d\n", port)
	return http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
