package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// readCommit loads a commit by id from objects/commits/<shard>/<rest>.
func readCommit(root, id string) (*Commit, error) {
	dir := filepath.Join(root, "objects", "commits", id[:2])
	path := filepath.Join(dir, id[2:])
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	r := bufio.NewReader(strings.NewReader(string(b)))

	read := func() (string, error) {
		s, err := r.ReadString('\n')
		if err != nil && !errors.Is(err, io.EOF) {
			return "", err
		}
		return strings.TrimRight(s, "\n"), err
	}

	expect := func(got, want string) error {
		if got != want {
			return fmt.Errorf("bad commit: expected %q, got %q", want, got)
		}
		return nil
	}

	// Follows CanonicalBytes(): key\nvalue\n ... then "files\n" then entries.
	c := &Commit{Files: map[string]string{}}

	l, _ := read(); if err := expect(l, "message"); err != nil { return nil, err }
	c.Message, _ = read()
	l, _ = read(); if err := expect(l, "timestamp"); err != nil { return nil, err }
	c.TimestampRFC, _ = read()
	l, _ = read(); if err := expect(l, "parent"); err != nil { return nil, err }
	c.Parent, _ = read()
	l, _ = read(); if err := expect(l, "parent2"); err != nil { return nil, err }
	c.SecondParent, _ = read()
	l, _ = read(); if err := expect(l, "files"); err != nil { return nil, err }

	for {
		line, err2 := read()
		if line == "" && errors.Is(err2, io.EOF) {
			break
		}
		if line == "" && err2 == nil {
			continue
		}
		parts := strings.SplitN(line, "\t", 2)
		if len(parts) == 2 {
			c.Files[parts[0]] = parts[1]
		}
		if errors.Is(err2, io.EOF) {
			break
		}
	}
	return c, nil
}
