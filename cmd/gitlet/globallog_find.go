package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

func printCommitEntry(id string, c *Commit) {
	fmt.Println("===")
	fmt.Printf("commit %s\n", id)
	if c.SecondParent != "" && len(c.Parent) >= 7 && len(c.SecondParent) >= 7 {
		fmt.Printf("Merge: %s %s\n", c.Parent[:7], c.SecondParent[:7])
	}
	tt, _ := time.Parse(time.RFC3339, c.TimestampRFC)
	local := tt.In(time.Local)
	fmt.Printf("Date: %s\n", local.Format("Mon Jan _2 15:04:05 2006 -0700"))
	fmt.Println(c.Message)
	fmt.Println()
}

// Walk all commit objects under .gitlet/objects/commits/** and print them.
// Order doesn't matter, but we'll sort IDs for stability.
func GlobalLogCmd(cwd string) error {
	root, err := gitRoot(cwd)
	if err != nil { return errNotRepo }

	var ids []string
	base := filepath.Join(root, "objects", "commits")
	shards, _ := os.ReadDir(base)
	for _, sh := range shards {
		if !sh.IsDir() { continue }
		dir := filepath.Join(base, sh.Name())
		files, _ := os.ReadDir(dir)
		for _, f := range files {
			if f.IsDir() { continue }
			ids = append(ids, sh.Name()+f.Name())
		}
	}
	sort.Strings(ids)

	for _, id := range ids {
		c, err := readCommit(root, id)
		if err == nil {
			printCommitEntry(id, c)
		}
	}
	return nil
}

func FindCmd(cwd, msg string) error {
	root, err := gitRoot(cwd)
	if err != nil { return errNotRepo }

	found := false
	base := filepath.Join(root, "objects", "commits")
	shards, _ := os.ReadDir(base)
	for _, sh := range shards {
		if !sh.IsDir() { continue }
		dir := filepath.Join(base, sh.Name())
		files, _ := os.ReadDir(dir)
		for _, f := range files {
			if f.IsDir() { continue }
			id := sh.Name()+f.Name()
			c, err := readCommit(root, id)
			if err == nil && c.Message == msg {
				fmt.Println(id)
				found = true
			}
		}
	}
	if !found {
		fmt.Println("Found no commit with that message.")
	}
	return nil
}
