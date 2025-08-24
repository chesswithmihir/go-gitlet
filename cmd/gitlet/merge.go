// cmd/gitlet/merge.go
package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func MergeCmd(cwd, otherBranch string) error {
	root, err := gitRoot(cwd)
	if err != nil { return errNotRepo }

	// branch exists?
	otherID, err := readBranchID(root, otherBranch)
	if err != nil { return errors.New("A branch with that name does not exist.") }

	// self-merge?
	currRef, err := headRefPath(root)
	if err != nil { return err }
	currBranch := filepathBase(currRef)
	if otherBranch == currBranch {
		return errors.New("Cannot merge a branch with itself.")
	}

	// uncommitted changes?
	idx, _ := loadIndex(root)
	if len(idx.Adds) > 0 || len(idx.Removes) > 0 {
		return errors.New("You have uncommitted changes.")
	}

	// ids & commits
	currID, err := headCommitID(root); if err != nil { return err }
	curr, err := readCommit(root, currID); if err != nil { return err }
	other, err := readCommit(root, otherID); if err != nil { return err }

	// split point
	spID, err := splitPoint(root, currID, otherID); if err != nil { return err }
	if spID == otherID {
		printlnExact("Given branch is an ancestor of the current branch.")
		return nil
	}
	if spID == currID {
		if err := ResetCmd(cwd, otherID); err != nil { return err }
		printlnExact("Current branch fast-forwarded.")
		return nil
	}
	sp, err := readCommit(root, spID); if err != nil { return err }

	// ---------- Decide actions per file ----------
	type action struct {
		write bool   // write/replace with this blob
		del   bool   // delete
		bid   string // blob to write (for write)
		conf  bool   // was conflict content synthesized
	}

	// union of filenames across sp, curr, other
	union := map[string]struct{}{}
	for f := range sp.Files { union[f] = struct{}{} }
	for f := range curr.Files { union[f] = struct{}{} }
	for f := range other.Files { union[f] = struct{}{} }

	planned := map[string]action{}
	encounteredConflict := false

	eq := func(a, b string) bool { return a == b }
	modSince := func(now, base string) bool { return now != base }

	for f := range union {
		spB := sp.Files[f]
		curB := curr.Files[f]
		givB := other.Files[f]

		curMod := modSince(curB, spB)
		givMod := modSince(givB, spB)

		switch {
		// same change or both removed: no-op
		case eq(curB, givB):
			// nothing

		// modified in given only -> take given
		case givMod && !curMod:
			planned[f] = action{write: true, bid: givB}

		// modified in current only -> keep current
		case curMod && !givMod:
			// no-op

		// present at split, unmodified in current, absent in given -> remove
		case spB != "" && curB == spB && givB == "":
			planned[f] = action{del: true}

		// present at split, unmodified in given, absent in current -> remain absent
		case spB != "" && givB == spB && curB == "":
			// no-op

		// not at split; only in given -> add
		case spB == "" && givB != "" && curB == "":
			planned[f] = action{write: true, bid: givB}

		// not at split; only in current -> keep
		case spB == "" && curB != "" && givB == "":
			// no-op

		default:
			// conflict: synthesize content and stage it
			curData := []byte{}
			givData := []byte{}
			if curB != "" {
				if d, err := readBlob(root, curB); err == nil { curData = d }
			}
			if givB != "" {
				if d, err := readBlob(root, givB); err == nil { givData = d }
			}
			conf := []byte("<<<<<<< HEAD\n")
			conf = append(conf, curData...)
			conf = append(conf, []byte("\n=======\n")...)
			conf = append(conf, givData...)
			conf = append(conf, []byte("\n>>>>>>>\n")...)

			bid := blobID(conf)
			if err := ensureBlobStored(root, bid, conf); err != nil { return err }
			planned[f] = action{write: true, bid: bid, conf: true}
			encounteredConflict = true
		}
	}

	// ---------- Pre-check: untracked file in the way ----------
	for f, act := range planned {
		if !act.write { continue }
		abs := filepath.Join(cwd, f)
		if _, err := os.Stat(abs); err == nil {
			_, trackedNow := curr.Files[f]
			// idx is empty (we checked), so "untracked" = !trackedNow
			if !trackedNow {
				data, rerr := os.ReadFile(abs)
				if rerr != nil || blobID(data) != act.bid {
					return errors.New("There is an untracked file in the way; delete it, or add and commit it first.")
				}
			}
		}
	}

	// ---------- Apply to working dir + build new snapshot ----------
	newSnap := make(map[string]string, len(curr.Files))
	for k, v := range curr.Files { newSnap[k] = v }

	for f, act := range planned {
		if act.del {
			_ = os.Remove(filepath.Join(cwd, f))
			delete(newSnap, f)
		} else if act.write {
			data, err := readBlob(root, act.bid)
			if err != nil { return err }
			if err := os.WriteFile(filepath.Join(cwd, f), data, 0o644); err != nil {
				return err
			}
			newSnap[f] = act.bid
		}
	}

	// If nothing changed, echo the normal commit error
	if equalSnapshots(newSnap, curr.Files) {
		return errors.New("No changes added to the commit.")
	}

	// ---------- Write merge commit (two parents) ----------
	msg := fmt.Sprintf("Merged %s into %s.", otherBranch, currBranch)
	c := &Commit{
		Message:      msg,
		TimestampRFC: nowRFC3339UTC(),
		Parent:       currID,
		SecondParent: otherID,
		Files:        newSnap,
	}
	cid := c.ID()
	dir := filepath.Join(root, "objects", "commits", cid[:2])
	path := filepath.Join(dir, cid[2:])
	if err := os.MkdirAll(dir, 0o755); err != nil { return err }
	if err := writeAtomic(path, c.CanonicalBytes()); err != nil { return err }

	// advance current branch ref
	if err := writeAtomic(currRef, []byte(cid+"\n")); err != nil { return err }

	// clear index (merge auto-staged then committed)
	idx.clear()
	if err := idx.save(root); err != nil { return err }

	if encounteredConflict {
		printlnExact("Encountered a merge conflict.")
	}
	return nil
}

// helpers to keep imports minimal
func filepathBase(p string) string {
	i := strings.LastIndex(p, "/")
	if i < 0 { return p }
	return p[i+1:]
}

func printlnExact(s string) { println(s) }

func equalSnapshots(a, b map[string]string) bool {
	if len(a) != len(b) { return false }
	for k, v := range a {
		if b[k] != v { return false }
	}
	return true
}

// nowRFC3339UTC returns the current time in RFC3339 format in UTC.
func nowRFC3339UTC() string {
	return time.Now().UTC().Format(time.RFC3339)
}
