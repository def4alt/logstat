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

		e, skipped, err := parser.ProcessLog(file, config.Strict)
		if err != nil {
			return err
		}

		if skipped > 0 {
			fmt.Fprintf(os.Stderr, "Malformed lines skipped: %d\n", skipped)
		}

		entries = e
	} else {
		e, skipped, err := parser.ProcessLog(os.Stdin, config.Strict)
		if err != nil {
			return err
		}

		if skipped > 0 {
			fmt.Fprintf(os.Stderr, "Malformed lines skipped: %d\n", skipped)
		}

		entries = e
	}

	topk := stats.TopKHosts(entries, config.TopK)

	fmt.Printf("Total entries: %d\n", len(entries))

	fmt.Printf("Top %d entries:\n", config.TopK)
	for _, kv := range topk {
		fmt.Printf("%s: %d\n", kv.Key, kv.Value)
	}

	return nil
}
