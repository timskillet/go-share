package main

import (
	"log"
	"net/http"

	"github.com/timskillet/go-share/internal/tracker"
)

func main() {
	t := tracker.NewTracker()

	http.HandleFunc("/announce", t.Announce)
	http.HandleFunc("/peers", t.GetPeers)

	log.Println("Tracker running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
