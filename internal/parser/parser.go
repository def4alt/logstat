package parser

import (
	"bufio"
	"fmt"
	"io"
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

// host ident user [timestamp] "method path (protocol)" status bytes
func processLine(line string) (LogEntry, error) {
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

	if inQuotes || inBrackets {
		return LogEntry{}, fmt.Errorf("unbalanced quotes or brackets")
	}

	if current < 7 {
		return LogEntry{}, fmt.Errorf("too few fields: got %d, want 7", current)
	}

	requestParts := strings.SplitN(parts[4], " ", 3)
	if len(requestParts) < 2 {
		return LogEntry{}, fmt.Errorf("invalid request: %q", parts[4])
	}

	method := requestParts[0]
	path := requestParts[1]
	protocol := ""

	if len(requestParts) > 2 {
		protocol = requestParts[2]
	}

	entry := LogEntry{
		Host:      parts[0],
		Ident:     parts[1],
		User:      parts[2],
		Timestamp: parts[3],
		Method:    method,
		Path:      path,
		Protocol:  protocol,
		Status:    parts[5],
		Bytes:     parts[6],
	}

	if len(entry.Status) != 3 || entry.Status[0] < '1' || entry.Status[0] > '5' {
		return LogEntry{}, fmt.Errorf("invalid status code: %q", entry.Status)
	}

	return entry, nil
}

func ProcessLog(file io.Reader, strict bool) ([]LogEntry, int, error) {
	scanner := bufio.NewScanner(file)

	var entries []LogEntry
	skipped := 0
	total := 0

	for scanner.Scan() {
		line := scanner.Text()
		total++

		entry, err := processLine(line)
		if err != nil {
			if strict {
				return nil, skipped, fmt.Errorf("malformed entry at %d, aborting: %v", total, err)
			}

			skipped++
			continue
		}

		entries = append(entries, entry)
	}

	if err := scanner.Err(); err != nil {
		return nil, skipped, err
	}

	return entries, skipped, nil
}
