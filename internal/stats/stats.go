package stats

import (
	"sort"

	"github.com/def4alt/logstat/internal/parser"
	"github.com/def4alt/logstat/internal/types"
)

func TopKFrequent(entries []parser.LogEntry, k int) []types.KV[int] {
	m := make(map[string]int)
	for _, entry := range entries {
		m[entry.Host]++
	}

	var sorted []types.KV[int]
	for host, count := range m {
		sorted = append(sorted, types.KV[int]{Key: host, Value: count})
	}

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Value > sorted[j].Value
	})

	if len(sorted) > k {
		sorted = sorted[:k]
	}

	return sorted
}
