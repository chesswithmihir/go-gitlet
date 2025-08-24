// cmd/gitlet/commit.go
package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func CommitCmd(cwd, msg string) error {
	if strings.TrimSpace(msg) == "" {
		return errors.New("Please enter a commit message.")
	}
	root, err := gitRoot(cwd)
	if err != nil {
		return errNotRepo
	}

	// Load index (staged adds/removes)
	idx, err := loadIndex(root)
	if err != nil {
		return err
	}
	if len(idx.Adds) == 0 && len(idx.Removes) == 0 {
		return errors.New("No changes added to the commit.")
	}

	// Parent = current HEAD
	parentID, err := headCommitID(root)
	if err != nil {
		return err
	}
	parent, err := readCommit(root, parentID)
	if err != nil {
		return err
	}

	// Snapshot = copy of parent, then apply removes and adds
	newSnap := make(map[string]string, len(parent.Files))
	for k, v := range parent.Files {
		newSnap[k] = v
	}
	for f := range idx.Removes {
		delete(newSnap, f)
	}
	for f, bid := range idx.Adds {
		newSnap[f] = bid
	}

	// New commit
	c := &Commit{
		Message:      msg,
		TimestampRFC: time.Now().UTC().Format(time.RFC3339),
		Parent:       parentID,
		SecondParent: "",
		Files:        newSnap,
	}
	cid := c.ID()

	// Store commit object
	dir := filepath.Join(root, "objects", "commits", cid[:2])
	path := filepath.Join(dir, cid[2:])
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	if err := writeAtomic(path, c.CanonicalBytes()); err != nil {
		return err
	}

	// Move current branch ref to new commit
	refPath, err := headRefPath(root)
	if err != nil {
		return err
	}
	if err := writeAtomic(refPath, []byte(cid+"\n")); err != nil {
		return err
	}

	// Clear index
	idx.clear()
	return idx.save(root)
}
