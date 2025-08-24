// cmd/gitlet/clear.go

package main

import (
	"fmt"
	"os"
	"path/filepath"
)

// Clear removes all files in the current Gitlet repository, including the .gitlet directory itself. 
// it can be thought of as a "reset" for the entire repository. `rm -rf .gitlet` is basically what it does.

func Clear(cwd string) error {
	gitletDir := filepath.Join(cwd, ".gitlet")
	// Check if .gitlet directory exists
	if _, err := os.Stat(gitletDir); os.IsNotExist(err) {
		return fmt.Errorf("A Gitlet version-control system does not exist in the current directory.")
	}
	// Remove the .gitlet directory and all its contents
	if err := os.RemoveAll(gitletDir); err != nil {
		return fmt.Errorf("Failed to clear the Gitlet repository: %v", err)
	}
	return nil
}