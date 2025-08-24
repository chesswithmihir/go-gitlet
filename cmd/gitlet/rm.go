package main

import (
	"errors"
	"os"
	"path/filepath"
)

func RmCmd(cwd, filename string) error {
	root, err := gitRoot(cwd)
	if err != nil { return errNotRepo }

	// load index
	idx, err := loadIndex(root)
	if err != nil { return err }

	// load HEAD commit
	headID, err := headCommitID(root)
	if err != nil { return err }
	head, err := readCommit(root, headID)
	if err != nil { return err }

	_, stagedAdd := idx.Adds[filename]
	_, tracked := head.Files[filename]

	if !stagedAdd && !tracked {
		return errors.New("No reason to remove the file.")
	}

	// unstage addition if present
	delete(idx.Adds, filename)

	// if tracked: stage removal + delete from working dir if exists
	if tracked {
		idx.Removes[filename] = struct{}{}
		_ = os.Remove(filepath.Join(cwd, filename)) // ignore if already gone
	}

	return idx.save(root)
}
