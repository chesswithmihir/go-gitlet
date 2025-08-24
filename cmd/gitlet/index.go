package main

import (
	"bufio"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Index struct {
	Adds    map[string]string   // filename -> blobID
	Removes map[string]struct{} // set
}

func newIndex() *Index {
	return &Index{
		Adds:    map[string]string{},
		Removes: map[string]struct{}{},
	}
}

func indexPath(root string) string { return filepath.Join(root, "index") }

func loadIndex(root string) (*Index, error) {
	idx := newIndex()
	b, err := os.ReadFile(indexPath(root))
	if err != nil {
		return idx, nil // treat missing as empty index
	}
	sc := bufio.NewScanner(strings.NewReader(string(b)))
	for sc.Scan() {
		line := sc.Text()
		if strings.HasPrefix(line, "A\t") {
			p := strings.SplitN(line, "\t", 3)
			if len(p) == 3 {
				idx.Adds[p[1]] = p[2]
			}
		} else if strings.HasPrefix(line, "R\t") {
			p := strings.SplitN(line, "\t", 2)
			if len(p) == 2 {
				idx.Removes[p[1]] = struct{}{}
			}
		}
	}
	return idx, nil
}

func (i *Index) save(root string) error {
	var lines []string
	rm := make([]string, 0, len(i.Removes))
	for f := range i.Removes {
		rm = append(rm, f)
	}
	sort.Strings(rm)
	for _, f := range rm {
		lines = append(lines, "R\t"+f)
	}
	add := make([]string, 0, len(i.Adds))
	for f := range i.Adds {
		add = append(add, f)
	}
	sort.Strings(add)
	for _, f := range add {
		lines = append(lines, "A\t"+f+"\t"+i.Adds[f])
	}
	return writeAtomic(indexPath(root), []byte(strings.Join(lines, "\n")))
}

func (i *Index) clear() {
	i.Adds = map[string]string{}
	i.Removes = map[string]struct{}{}
}
