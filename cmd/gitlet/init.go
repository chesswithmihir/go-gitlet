// cmd/gitlet/init.go

package main

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// ---- Types and helpers for canonical, deterministic commits ----

type Commit struct {
	Message      string
	TimestampRFC string // store in RFC3339 UTC; we'll format for `log` later.
	Parent       string // empty for initial commit
	SecondParent string // empty unless merge
	Files        map[string]string // filename -> blobID (empty map for initial)
}

// CanonicalBytes builds a stable, language-agnostic byte layout.
// Sort filenames so Go's random map order can't change commit IDs.
func (c *Commit) CanonicalBytes() []byte {
	var b []byte
	appendKV := func(k, v string) {
		b = append(b, k...)
		b = append(b, '\n')
		b = append(b, v...)
		b = append(b, '\n')
	}
	appendKV("message", c.Message)
	appendKV("timestamp", c.TimestampRFC)
	appendKV("parent", c.Parent)
	appendKV("parent2", c.SecondParent)

	// Files section: sorted (filename, blobID) lines.
	b = append(b, "files\n"...)
	if len(c.Files) > 0 {
		keys := make([]string, 0, len(c.Files))
		for k := range c.Files {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			line := fmt.Sprintf("%s\t%s\n", k, c.Files[k])
			b = append(b, line...)
		}
	} else {
		// still add nothing beyond the header; the trailing newline is already there
	}
	return b
}

func (c *Commit) ID() string {
	h := sha1.New()
	h.Write([]byte("commit\n"))        // type tag to avoid blob/commit collisions
	h.Write(c.CanonicalBytes())        // canonical payload
	return hex.EncodeToString(h.Sum(nil))
}

// ---- Filesystem helpers (safe paths, atomic writes, etc.) ----

func writeAtomic(path string, data []byte) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, ".tmp-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	_, werr := tmp.Write(data)
	cerr := tmp.Close()
	if werr != nil {
		os.Remove(tmpName)
		return werr
	}
	if cerr != nil {
		os.Remove(tmpName)
		return cerr
	}
	// If destination already exists and has identical bytes, this will replace it,
	// which is fine (objects are immutable; refs can be updated).
	return os.Rename(tmpName, path)
}

func objectPath(root, kind, id string) (string, string) {
	// shard: ab/cdef...
	subdir := filepath.Join(root, "objects", kind, id[:2])
	return subdir, filepath.Join(subdir, id[2:])
}

// ---- Init command ----

func Init(cwd string) error {
	root := filepath.Join(cwd, ".gitlet")

	// Failure case: already initialized.
	if st, err := os.Stat(root); err == nil && st.IsDir() {
		return errors.New("A Gitlet version-control system already exists in the current directory.")
	}

	// Create base directories.
	dirs := []string{
		root,
		filepath.Join(root, "refs", "heads"),
		filepath.Join(root, "objects", "blobs"),
		filepath.Join(root, "objects", "commits"),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return err
		}
	}

	// Build the epoch initial commit: empty snapshot, fixed message, epoch time.
	initial := &Commit{
		Message:      "initial commit",
		TimestampRFC: time.Unix(0, 0).UTC().Format(time.RFC3339), // 1970-01-01T00:00:00Z
		Parent:       "",
		SecondParent: "",
		Files:        map[string]string{},
	}
	cid := initial.ID()

	// Write the commit object (id-sharded path).
	cdir, cpath := objectPath(root, "commits", cid)
	if err := os.MkdirAll(cdir, 0o755); err != nil {
		return err
	}
	// Avoid rewriting if it somehow exists (idempotent).
	if _, err := os.Stat(cpath); os.IsNotExist(err) {
		if err := writeAtomic(cpath, initial.CanonicalBytes()); err != nil {
			return err
		}
	}

	// Write HEAD (symbolic ref) and master tip.
	if err := writeAtomic(filepath.Join(root, "HEAD"), []byte("ref: refs/heads/master\n")); err != nil {
		return err
	}
	if err := writeAtomic(filepath.Join(root, "refs", "heads", "master"), []byte(cid+"\n")); err != nil {
		return err
	}

	// Success: per spec, print nothing.
	return nil
}
