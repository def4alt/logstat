package parser

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

func ProcessLog(file io.Reader) {
	fmt.Println("Processing log file...")

	scanner := bufio.NewScanner(file)

	m := make(map[string]int)

	for scanner.Scan() {
		line := scanner.Text()

		parts := strings.SplitN(line, " ", 2)

		if len(parts) < 2 {
			continue
		}

		key := parts[0]

		m[key]++
	}

	fmt.Println("Summary statistics:")
	for key, count := range m {
		fmt.Printf("%s: %d\n", key, count)
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading log file: %v\n", err)
	}
}
