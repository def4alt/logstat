package stats

import (
	"sort"
	"strconv"

	"github.com/def4alt/logstat/internal/parser"
	"github.com/def4alt/logstat/internal/types"
)

func numberOrZero(s string) int {
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}

	return n
}

func TotalEntries(entries []parser.LogEntry) int {
	return len(entries)
}

func TotalBytes(entries []parser.LogEntry) int {
	total := 0

	for _, entry := range entries {
		bytes := numberOrZero(entry.Bytes)
		total += bytes
	}

	return total
}

func UniqueHosts(entries []parser.LogEntry) int {
	hosts := make(map[string]struct{})
	for _, entry := range entries {
		hosts[entry.Host] = struct{}{}
	}
	return len(hosts)
}

func StatusCodeCounts(entries []parser.LogEntry) map[string]int {
	counts := make(map[string]int)
	for _, entry := range entries {
		counts[entry.Status]++
	}
	return counts
}

func MethodCounts(entries []parser.LogEntry) map[string]int {
	counts := make(map[string]int)
	for _, entry := range entries {
		counts[entry.Method]++
	}
	return counts
}

func topK(entries []parser.LogEntry, keyFn func(parser.LogEntry) string, k int) []types.KV[int] {
	m := make(map[string]int)
	for _, e := range entries {
		m[keyFn(e)]++
	}

	var sorted []types.KV[int]
	for path, count := range m {
		sorted = append(sorted, types.KV[int]{Key: path, Value: count})
	}

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Value > sorted[j].Value
	})

	if len(sorted) > k {
		sorted = sorted[:k]
	}

	return sorted
}

func TopKPaths(entries []parser.LogEntry, k int) []types.KV[int] {
	return topK(entries, func(e parser.LogEntry) string { return e.Path }, k)
}

func TopKHosts(entries []parser.LogEntry, k int) []types.KV[int] {
	return topK(entries, func(e parser.LogEntry) string { return e.Host }, k)
}

func percentile(entries []parser.LogEntry, p float64) int {
	var bytes []int

	for _, entry := range entries {
		n := numberOrZero(entry.Bytes)
		bytes = append(bytes, n)
	}

	sort.Ints(bytes)

	if len(bytes) == 0 {
		return 0
	}

	index := int(float64(len(bytes)) * p)
	if index >= len(bytes) {
		index = len(bytes) - 1
	}

	return bytes[index]
}

func P50Bytes(entries []parser.LogEntry) int {
	return percentile(entries, 0.5)
}

func P90Bytes(entries []parser.LogEntry) int {
	return percentile(entries, 0.9)
}

func P95Bytes(entries []parser.LogEntry) int {
	return percentile(entries, 0.95)
}

func P99Bytes(entries []parser.LogEntry) int {
	return percentile(entries, 0.99)
}
