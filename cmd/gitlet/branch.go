package main

import (
	"errors"
	"os"
	"path/filepath"
)

func BranchCmd(cwd, name string) error {
	root, err := gitRoot(cwd)
	if err != nil { return errNotRepo }

	refPath := filepath.Join(root, "refs", "heads", name)
	if _, err := os.Stat(refPath); err == nil {
		return errors.New("A branch with that name already exists.")
	}

	headID, err := headCommitID(root)
	if err != nil { return err }

	return writeAtomic(refPath, []byte(headID+"\n"))
}

func RmBranchCmd(cwd, name string) error {
	root, err := gitRoot(cwd)
	if err != nil { return errNotRepo }

	refPath := filepath.Join(root, "refs", "heads", name)
	if _, err := os.Stat(refPath); err != nil {
		return errors.New("A branch with that name does not exist.")
	}

	currRef, err := headRefPath(root)
	if err != nil { return err }
	curr := filepath.Base(currRef)
	if name == curr {
		return errors.New("Cannot remove the current branch.")
	}

	return os.Remove(refPath)
}
