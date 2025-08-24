package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// .gitlet/HEAD contains: "ref: refs/heads/<name>\n"
func headRefPath(root string) (string, error) {
	b, err := os.ReadFile(filepath.Join(root, "HEAD"))
	if err != nil {
		return "", err
	}
	line := strings.TrimSpace(string(b))
	const pfx = "ref: "
	if !strings.HasPrefix(line, pfx) {
		return "", fmt.Errorf("HEAD not a symbolic ref")
	}
	rel := strings.TrimSpace(line[len(pfx):])
	return filepath.Join(root, filepath.FromSlash(rel)), nil
}

func headCommitID(root string) (string, error) {
	ref, err := headRefPath(root)
	if err != nil {
		return "", err
	}
	b, err := os.ReadFile(ref)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(b)), nil
}
