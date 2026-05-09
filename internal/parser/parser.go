package parser

import (
	"bufio"
	"fmt"
	"io"
	"sort"
	"strings"
)

func ProcessLog(file io.Reader, topk int) {
	fmt.Println("Processing log file...")

	scanner := bufio.NewScanner(file)

	m := make(map[string]int)
	total := 0

	for scanner.Scan() {
		line := scanner.Text()

		parts := strings.SplitN(line, " ", 2)

		if len(parts) < 2 {
			continue
		}

		key := parts[0]

		m[key]++
		total++
	}

	type kv struct {
		Key   string
		Value int
	}

	var sorted []kv
	for k, v := range m {
		sorted = append(sorted, kv{k, v})
	}

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Value > sorted[j].Value
	})

	if len(sorted) > topk {
		sorted = sorted[:topk]
	}

	fmt.Println("Summary statistics:")

	fmt.Printf("Total entries: %d\n", total)

	fmt.Printf("Top %d entries:\n", topk)
	for i := 0; i < topk && i < len(sorted); i++ {
		fmt.Printf("%s: %d\n", sorted[i].Key, sorted[i].Value)
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading log file: %v\n", err)
	}
}
