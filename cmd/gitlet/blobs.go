package main

import (
	"crypto/sha1"
	"encoding/hex"
	"os"
	"path/filepath"
)

// blobID = SHA1("blob\n" + file bytes)
func blobID(data []byte) string {
	h := sha1.New()
	h.Write([]byte("blob\n"))
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}

// ensureBlobStored writes the blob object if it doesn't already exist.
func ensureBlobStored(root, id string, data []byte) error {
	dir := filepath.Join(root, "objects", "blobs", id[:2])
	path := filepath.Join(dir, id[2:])
	if _, err := os.Stat(path); err == nil {
		return nil // already there
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	return writeAtomic(path, data)
}

func readBlob(root, id string) ([]byte, error) {
	dir := filepath.Join(root, "objects", "blobs", id[:2])
	path := filepath.Join(dir, id[2:])
	return os.ReadFile(path)
}
