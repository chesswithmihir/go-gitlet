package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

// resolveCommitID returns a full 40-hex id for a given prefix (or exact id).
// If no unique match exists, returns "No commit with that id exists."
func resolveCommitID(root, prefix string) (string, error) {
	prefix = strings.TrimSpace(prefix)
	if len(prefix) == 40 {
		// verify it exists on disk
		dir := filepath.Join(root, "objects", "commits", prefix[:2])
		path := filepath.Join(dir, prefix[2:])
		if _, err := os.Stat(path); err == nil {
			return prefix, nil
		}
		return "", errors.New("No commit with that id exists.")
	}

	commitsDir := filepath.Join(root, "objects", "commits")
	var subdirs []string

	// If we have >=2 hex, we only need to search that shard.
	if len(prefix) >= 2 {
		sub := filepath.Join(commitsDir, prefix[:2])
		if fi, err := os.Stat(sub); err == nil && fi.IsDir() {
			subdirs = []string{sub}
		}
	} else {
		entries, _ := os.ReadDir(commitsDir)
		for _, e := range entries {
			if e.IsDir() {
				subdirs = append(subdirs, filepath.Join(commitsDir, e.Name()))
			}
		}
	}

	match := ""
	want := prefix
	for _, d := range subdirs {
		entries, _ := os.ReadDir(d)
		for _, e := range entries {
			id := filepath.Base(d) + e.Name() // full id = shard + rest
			if strings.HasPrefix(id, want) {
				if match == "" {
					match = id
				} else if match != id {
					// ambiguous; spec doesn’t give a special message, we’ll just fail
					return "", errors.New("No commit with that id exists.")
				}
			}
		}
	}
	if match == "" {
		return "", errors.New("No commit with that id exists.")
	}
	return match, nil
}
