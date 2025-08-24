package main

import (
	"fmt"
	"time"
)

// spec format example:
// ===
// commit <40-hex>
// Merge: 4975af1 2c1ead1         // only for merge commits
// Date: Thu Nov 9 20:00:05 2017 -0800
// <message>
// 
func LogCmd(cwd string) error {
	root, err := gitRoot(cwd)
	if err != nil {
		return errNotRepo
	}
	id, err := headCommitID(root)
	if err != nil {
		return err
	}
	for id != "" {
		c, err := readCommit(root, id)
		if err != nil {
			return err
		}
		fmt.Println("===")
		fmt.Printf("commit %s\n", id)

		// Merge line (only if SecondParent set)
		if c.SecondParent != "" && len(c.Parent) >= 7 && len(c.SecondParent) >= 7 {
			fmt.Printf("Merge: %s %s\n", c.Parent[:7], c.SecondParent[:7])
		}

		// parse stored RFC3339 (UTC), show in local time with spec layout
		tt, _ := time.Parse(time.RFC3339, c.TimestampRFC)
		local := tt.In(time.Local)
		fmt.Printf("Date: %s\n", local.Format("Mon Jan _2 15:04:05 2006 -0700"))

		fmt.Println(c.Message)
		fmt.Println() // blank line after each entry

		id = c.Parent
	}
	return nil
}
