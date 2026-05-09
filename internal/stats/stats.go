package stats

import (
	"sort"

	"github.com/def4alt/logstat/internal/parser"
)

type kv struct {
	Key   string
	Value int
}

func TopKFrequent(entries []parser.LogEntry, k int) []kv {
	m := make(map[string]int)
	for _, entry := range entries {
		m[entry.Host]++
	}

	var sorted []kv
	for k, v := range m {
		sorted = append(sorted, kv{k, v})
	}

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Value > sorted[j].Value
	})

	if len(sorted) > k {
		sorted = sorted[:k]
	}

	return sorted
}
