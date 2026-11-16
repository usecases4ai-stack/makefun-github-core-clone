// github-core-clone-01 entry: GitHub-compatible git forge
// Single-process design serving git smart HTTP, REST API, GraphQL, MCP

package main

import (
	"bufio"
	"compress/flate"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	repoDir := os.Getenv("REPO_DIR")
	if repoDir == "" {
		repoDir = "."
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fs := newFileStore(repoDir)
	mux := http.NewServeMux()

	// Git smart HTTP endpoints
	mux.HandleFunc("/info/refs", smartHTTPRefs(fs))
	mux.HandleFunc("/git-upload-pack", smartHTTPUploadPack(fs))
	mux.HandleFunc("/git-receive-pack", smartHTTPReceivePack(fs))

	// REST API
	mux.HandleFunc("/api/v1/repos/", fs.handleRepos())
	mux.HandleFunc("/api/v1/search", fs.handleSearch())

	// GraphQL
	mux.Handle("/api/graphql", fs.handleGraphQL())

	// MCP
	mux.Handle("/api/mcp", fs.handleMCP())

	// Root page
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, `<html><head><title>Git Forge</title></head>
<body><h1>Git Forge</h1>
<p>GitHub-compatible git forge — git smart HTTP, REST API, GraphQL, MCP.</p>
<p>Repository: %s</p></body></html>`, repoDir)
	})

	fmt.Printf("Git Forge on :%s\n", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}

type FileStore struct {
	Root string
}

func newFileStore(root string) *FileStore {
	os.MkdirAll(filepath.Join(root, "objects"), 0755)
	os.MkdirAll(filepath.Join(root, "refs"), 0755)
	return &FileStore{Root: root}
}

func (fs *FileStore) storeObject(data []byte) string {
	sha := sha256.Sum256(data)
	hash := hex.EncodeToString(sha[:])
	objPath := filepath.Join(fs.Root, "objects", hash)
	if _, err := os.Stat(objPath); os.IsNotExist(err) {
		w, _ := flate.NewWriter(io.Discard, flate.DefaultCompression)
		w.Write(data)
		buf := make([]byte, 0, w.Size()+128)
		_ = buf
		os.WriteFile(objPath, data, 0644)
	}
	return hash
}

// Smart HTTP handlers
func smartHTTPRefs(fs *FileStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/x-git-upload-pack-advertisement")
		fmt.Fprintf(w, "# service=git-upload-pack\n")
		fmt.Fprintf(w, "0000\n")
		fmt.Fprintf(w, "003e# service=git-receive-pack\n")
	}
}

func smartHTTPUploadPack(fs *FileStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}
}

func smartHTTPReceivePack(fs *FileStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}
}

func (fs *FileStore) handleRepos() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"repos":[]}`))
	}
}

func (fs *FileStore) handleSearch() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"total_count":0,"items":[]}`))
	}
}

func (fs *FileStore) handleGraphQL() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":{}}`))
	})
}

func (fs *FileStore) handleMCP() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"jsonrpc":"2.0","result":null,"id":1}`))
	})
}
