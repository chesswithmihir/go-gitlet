package main

import (
	"container/list"
	"os"
	"path/filepath"
	"strings"
)

// parents of a commit (0, 1, or 2)
func commitParents(c *Commit) []string {
	ps := []string{}
	if c.Parent != "" { ps = append(ps, c.Parent) }
	if c.SecondParent != "" { ps = append(ps, c.SecondParent) }
	return ps
}

// ancestorsMap returns map[id]distance (min edges from start back through parents).
func ancestorsMap(root, start string) (map[string]int, error) {
	dist := map[string]int{start: 0}
	q := list.New()
	q.PushBack(start)

	for q.Len() > 0 {
		id := q.Remove(q.Front()).(string)
		c, err := readCommit(root, id)
		if err != nil { return nil, err }
		for _, p := range commitParents(c) {
			if _, seen := dist[p]; !seen {
				dist[p] = dist[id] + 1
				q.PushBack(p)
			}
		}
	}
	return dist, nil
}

// splitPoint picks a common ancestor that is "latest" (closest to the heads).
// We approximate by minimizing distA+distB (ties broken by smaller distA).
func splitPoint(root, aID, bID string) (string, error) {
	da, err := ancestorsMap(root, aID); if err != nil { return "", err }
	db, err := ancestorsMap(root, bID); if err != nil { return "", err }
	best := ""
	bestSum := int(^uint(0) >> 1) // max int
	bestA := bestSum

	for id, d1 := range da {
		if d2, ok := db[id]; ok {
			sum := d1 + d2
			if sum < bestSum || (sum == bestSum && d1 < bestA) {
				best, bestSum, bestA = id, sum, d1
			}
		}
	}
	return best, nil
}

// tiny helper: read a branch ref to full id
func readBranchID(root, name string) (string, error) {
	b, err := os.ReadFile(filepath.Join(root, "refs", "heads", name))
	if err != nil { return "", err }
	return strings.TrimSpace(string(b)), nil
}
