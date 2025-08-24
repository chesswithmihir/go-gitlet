package main

import (
	"errors"
	"os"
	"path/filepath"
)

// CheckoutBranchCmd switches to <branch> per spec.
func CheckoutBranchCmd(cwd, branch string) error {
	root, err := gitRoot(cwd)
	if err != nil { return errNotRepo }

	// Branch must exist.
	targetRef := filepath.Join(root, "refs", "heads", branch)
	if _, err := os.Stat(targetRef); err != nil {
		return errors.New("No such branch exists.")
	}

	// Must not already be current branch.
	currRefPath, err := headRefPath(root)
	if err != nil { return err }
	if filepath.Base(currRefPath) == branch {
		return errors.New("No need to checkout the current branch.")
	}

	// Load commits.
	targetIDBytes, _ := os.ReadFile(targetRef)
	targetID := string(bytesTrimNL(targetIDBytes))
	target, err := readCommit(root, targetID)
	if err != nil { return err }

	currID, err := headCommitID(root)
	if err != nil { return err }
	curr, err := readCommit(root, currID)
	if err != nil { return err }

	// Load index to detect "untracked" (not tracked in curr and not staged for add).
	idx, _ := loadIndex(root)

	// Pre-check: untracked file that would be overwritten by checkout.
	for fname, bid := range target.Files {
		abs := filepath.Join(cwd, fname)
		if _, err := os.Stat(abs); err == nil {
			_, trackedNow := curr.Files[fname]
			_, stagedAdd := idx.Adds[fname]
			if !trackedNow && !stagedAdd {
				// Optional: compare contents to see if truly overwritten
				if data, err := os.ReadFile(abs); err == nil {
					if blobID(data) != bid {
						return errors.New("There is an untracked file in the way; delete it, or add and commit it first.")
					}
				} else {
					// Can't read; be conservative.
					return errors.New("There is an untracked file in the way; delete it, or add and commit it first.")
				}
			}
		}
	}

	// Write all files from target snapshot.
	for fname, bid := range target.Files {
		data, err := readBlob(root, bid)
		if err != nil { return err }
		if err := os.WriteFile(filepath.Join(cwd, fname), data, 0o644); err != nil {
			return err
		}
	}

	// Remove files tracked in current but not in target.
	for fname := range curr.Files {
		if _, ok := target.Files[fname]; !ok {
			_ = os.Remove(filepath.Join(cwd, fname))
		}
	}

	// Clear staging area.
	idx.clear()
	if err := idx.save(root); err != nil { return err }

	// Point HEAD to the branch.
	if err := writeAtomic(filepath.Join(root, "HEAD"), []byte("ref: refs/heads/"+branch+"\n")); err != nil {
		return err
	}
	return nil
}

// tiny helper: trim trailing newline from ref files
func bytesTrimNL(b []byte) []byte {
	if n := len(b); n > 0 && (b[n-1] == '\n' || b[n-1] == '\r') {
		return b[:n-1]
	}
	return b
}
