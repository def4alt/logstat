package parser

import (
	"bufio"
	"fmt"
	"io"
	"sort"
	"strings"
)

// host ident user [timestamp] "method path protocol" status bytes
func processLine(line string, strict bool) ([7]string, error) {
	inBrackets := false
	inQuotes := false

	var parts [7]string
	current := 0

	var builder strings.Builder

	for i := 0; i < len(line); i++ {
		c := line[i]

		if c == ' ' && !inBrackets && !inQuotes && current < len(parts) {
			parts[current] = builder.String()
			current++
			builder.Reset()
			continue
		}

		switch c {
		case '[':
			inBrackets = true
		case ']':
			inBrackets = false
		case '"':
			inQuotes = !inQuotes
		default:
			builder.WriteByte(c)
		}

	}

	if builder.Len() > 0 && current < len(parts) {
		parts[current] = builder.String()
		current++
	}

	if current < 7 && strict {
		return parts, fmt.Errorf("unexpected log format: %s", line)
	}

	return parts, nil
}

func ProcessLog(file io.Reader, topk int, strict bool) {
	fmt.Println("Processing log file...")

	scanner := bufio.NewScanner(file)

	m := make(map[string]int)
	total := 0

	for scanner.Scan() {
		line := scanner.Text()

		parts, err := processLine(line, strict)
		if err != nil {
			fmt.Printf("Warning: %v\n", err)
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
