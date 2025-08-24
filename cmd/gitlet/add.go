package main

import (
	"errors"
	"os"
	"path/filepath"
)

func Add(cwd, filename string) error {
	// Must be in a repo
	root, err := gitRoot(cwd)
	if err != nil {
		return errNotRepo
	}

	// File must exist
	abs := filepath.Join(cwd, filename)
	data, err := os.ReadFile(abs)
	if err != nil {
		return errors.New("File does not exist.")
	}

	// Compute blob id + store
	bid := blobID(data)
	if err := ensureBlobStored(root, bid, data); err != nil {
		return err
	}

	// Load HEAD commit to compare
	headID, err := headCommitID(root)
	if err != nil {
		return err
	}
	head, err := readCommit(root, headID)
	if err != nil {
		return err
	}
	headBlob := head.Files[filename] // "" if not tracked

	// Load index, update according to spec
	idx, err := loadIndex(root)
	if err != nil {
		return err
	}

	if headBlob == bid {
		// identical to HEAD: unstage add + unstage removal
		delete(idx.Adds, filename)
		delete(idx.Removes, filename)
	} else {
		// different: stage for add, unstage removal
		idx.Adds[filename] = bid
		delete(idx.Removes, filename)
	}

	return idx.save(root)
}
