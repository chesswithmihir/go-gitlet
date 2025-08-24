package main

import (
	"errors"
	"os"
	"path/filepath"
)

// ResetCmd: reset <commit-id/prefix>
func ResetCmd(cwd, prefix string) error {
	root, err := gitRoot(cwd)
	if err != nil { return errNotRepo }

	// Resolve target commit ID (supports abbreviated ids)
	cid, err := resolveCommitID(root, prefix)
	if err != nil { return err } // prints: No commit with that id exists.

	// Load target and current commits
	target, err := readCommit(root, cid)
	if err != nil { return err }

	curID, err := headCommitID(root)
	if err != nil { return err }
	current, err := readCommit(root, curID)
	if err != nil { return err }

	// Load index (to detect untracked files)
	idx, _ := loadIndex(root)

	// Pre-check: untracked files that would be overwritten by target
	for fname, bid := range target.Files {
		abs := filepath.Join(cwd, fname)
		if _, err := os.Stat(abs); err == nil {
			_, trackedNow := current.Files[fname]
			_, stagedAdd := idx.Adds[fname]
			if !trackedNow && !stagedAdd {
				if data, err := os.ReadFile(abs); err == nil {
					if blobID(data) != bid {
						return errors.New("There is an untracked file in the way; delete it, or add and commit it first.")
					}
				} else {
					return errors.New("There is an untracked file in the way; delete it, or add and commit it first.")
				}
			}
		}
	}

	// Write all files from target snapshot
	for fname, bid := range target.Files {
		data, err := readBlob(root, bid)
		if err != nil { return err }
		if err := os.WriteFile(filepath.Join(cwd, fname), data, 0o644); err != nil {
			return err
		}
	}

	// Remove files tracked now but absent in target
	for fname := range current.Files {
		if _, ok := target.Files[fname]; !ok {
			_ = os.Remove(filepath.Join(cwd, fname))
		}
	}

	// Clear index
	idx.clear()
	if err := idx.save(root); err != nil { return err }

	// Move current branch ref to target commit (HEAD stays pointing to this ref)
	refPath, err := headRefPath(root)
	if err != nil { return err }
	if err := writeAtomic(refPath, []byte(cid+"\n")); err != nil { return err }

	return nil
}
