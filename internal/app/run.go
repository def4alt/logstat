package app

import (
	"fmt"
	"os"

	"github.com/def4alt/logstat/internal/parser"
	"github.com/def4alt/logstat/internal/stats"
)

type Config struct {
	FilePath string
	JSON     bool
	TopK     int
	Strict   bool
}

func Run(config Config) error {
	var entries []parser.LogEntry
	var malformedEntries []parser.LogEntry

	if config.FilePath != "" {
		file, err := os.OpenFile(config.FilePath, os.O_RDONLY, 0o644)
		if err != nil {
			return err
		}
		defer func() {
			if cerr := file.Close(); cerr != nil && err == nil {
				err = cerr
			}
		}()

		e, me, err := parser.ProcessLog(file, config.Strict)
		if err != nil {
			return err
		}

		entries = e
		malformedEntries = me
	} else {
		e, me, err := parser.ProcessLog(os.Stdin, config.Strict)
		if err != nil {
			return err
		}

		entries = e
		malformedEntries = me
	}

	for _, entry := range malformedEntries {
		fmt.Printf("Malformed entry: %s\n", entry.Host)
	}

	topk := stats.TopKFrequent(entries, config.TopK)

	fmt.Printf("Total entries: %d\n", len(entries))
	fmt.Printf("Malformed entries: %d\n", len(malformedEntries))

	fmt.Printf("Top %d entries:\n", config.TopK)
	for _, kv := range topk {
		fmt.Printf("%s: %d\n", kv.Key, kv.Value)
	}

	return nil
}
