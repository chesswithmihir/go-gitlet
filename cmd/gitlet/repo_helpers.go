package main

import (
	"errors"
	"os"
	"path/filepath"
)

var errNotRepo = errors.New("Not in an initialized Gitlet directory.")

// gitRoot returns "<cwd>/.gitlet" if it exists, else errNotRepo.
func gitRoot(cwd string) (string, error) {
	root := filepath.Join(cwd, ".gitlet")
	st, err := os.Stat(root)
	if err == nil && st.IsDir() {
		return root, nil
	}
	return "", errNotRepo
}
