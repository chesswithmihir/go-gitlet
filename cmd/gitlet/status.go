package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

func StatusCmd(cwd string) error {
	root, err := gitRoot(cwd)
	if err != nil { return errNotRepo }

	// --- Branches ---
	headsDir := filepath.Join(root, "refs", "heads")
	ents, _ := os.ReadDir(headsDir)
	var branches []string
	for _, e := range ents {
		if !e.IsDir() { branches = append(branches, e.Name()) }
	}
	sort.Strings(branches)

	refPath, err := headRefPath(root)
	if err != nil { return err }
	curr := filepath.Base(refPath)

	fmt.Println("=== Branches ===")
	for _, b := range branches {
		if b == curr { fmt.Printf("*%s\n", b) } else { fmt.Println(b) }
	}
	fmt.Println()

	// --- Staged Files ---
	idx, _ := loadIndex(root)
	var adds []string
	for f := range idx.Adds { adds = append(adds, f) }
	sort.Strings(adds)
	fmt.Println("=== Staged Files ===")
	for _, f := range adds { fmt.Println(f) }
	fmt.Println()

	// --- Removed Files ---
	var rms []string
	for f := range idx.Removes { rms = append(rms, f) }
	sort.Strings(rms)
	fmt.Println("=== Removed Files ===")
	for _, f := range rms { fmt.Println(f) }
	fmt.Println()

	// --- Extra credit sections (leave empty for now) ---
	fmt.Println("=== Modifications Not Staged For Commit ===")
	fmt.Println()
	fmt.Println("=== Untracked Files ===")
	fmt.Println()

	return nil
}
