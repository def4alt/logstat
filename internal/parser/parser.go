package parser

import (
	"bufio"
	"fmt"
	"io"
	"sort"
	"strings"
)

type LogEntry struct {
	Host      string
	Ident     string
	User      string
	Timestamp string
	Method    string
	Path      string
	Protocol  string
	Status    string
	Bytes     string
}

// host ident user [timestamp] "method path protocol" status bytes
func processLine(line string) (LogEntry, error) {
	inBrackets := false

	var parts [9]string
	current := 0

	var builder strings.Builder

	for i := 0; i < len(line); i++ {
		c := line[i]

		if c == ' ' && !inBrackets && current < len(parts) {
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
		default:
			builder.WriteByte(c)
		}

	}

	if builder.Len() > 0 && current < len(parts) {
		parts[current] = builder.String()
		current++
	}

	entry := LogEntry{
		Host:      parts[0],
		Ident:     parts[1],
		User:      parts[2],
		Timestamp: parts[3],
		Method:    parts[4],
		Path:      parts[5],
		Protocol:  parts[6],
		Status:    parts[7],
		Bytes:     parts[8],
	}

	if current < 9 {
		return entry, fmt.Errorf("unexpected log format: %s", line)
	}

	return entry, nil
}

func ProcessLog(file io.Reader, topk int, strict bool) error {
	fmt.Println("Processing log file...")

	scanner := bufio.NewScanner(file)

	m := make(map[string]int)
	total := 0
	malformed := 0

	for scanner.Scan() {
		line := scanner.Text()

		entry, err := processLine(line)
		if err != nil {
			if strict {
				return fmt.Errorf("malformed entry, aborting: %v", err)
			}

			fmt.Printf("Warning: %v\n", err)
			malformed++
			continue
		}

		key := entry.Host

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
	fmt.Printf("Malformed entries: %d\n", malformed)

	fmt.Printf("Top %d entries:\n", topk)
	for i := 0; i < topk && i < len(sorted); i++ {
		fmt.Printf("%s: %d\n", sorted[i].Key, sorted[i].Value)
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading log file: %v\n", err)
	}

	return nil
}
