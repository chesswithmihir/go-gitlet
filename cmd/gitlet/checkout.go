package main

import (
	"errors"
	"os"
	"path/filepath"
)

// checkout -- <file>
func CheckoutHeadFile(cwd, filename string) error {
	root, err := gitRoot(cwd)
	if err != nil {
		return errNotRepo
	}
	headID, err := headCommitID(root)
	if err != nil {
		return err
	}
	c, err := readCommit(root, headID)
	if err != nil {
		return err
	}
	bid, ok := c.Files[filename]
	if !ok || bid == "" {
		return errors.New("File does not exist in that commit.")
	}
	data, err := readBlob(root, bid)
	if err != nil {
		return err
	}
	dest := filepath.Join(cwd, filename)
	return os.WriteFile(dest, data, 0o644)
}

// checkout <commit> -- <file>
func CheckoutCommitFile(cwd, commitPrefix, filename string) error {
	root, err := gitRoot(cwd)
	if err != nil {
		return errNotRepo
	}
	cid, err := resolveCommitID(root, commitPrefix)
	if err != nil {
		return err // prints "No commit with that id exists."
	}
	c, err := readCommit(root, cid)
	if err != nil {
		return err
	}
	bid, ok := c.Files[filename]
	if !ok || bid == "" {
		return errors.New("File does not exist in that commit.")
	}
	data, err := readBlob(root, bid)
	if err != nil {
		return err
	}
	dest := filepath.Join(cwd, filename)
	return os.WriteFile(dest, data, 0o644)
}

