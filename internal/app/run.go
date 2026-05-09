package app

import (
	"os"

	"github.com/def4alt/logstat/internal/output"
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
	var skipped int

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

		e, s, err := parser.ProcessLog(file, config.Strict)
		if err != nil {
			return err
		}

		entries = e
		skipped = s
	} else {
		e, s, err := parser.ProcessLog(os.Stdin, config.Strict)
		if err != nil {
			return err
		}

		entries = e
		skipped = s
	}

	summary := stats.GenerateSummary(entries, skipped, config.TopK)

	if config.JSON {
		return output.PrintJSON(summary)
	}

	return output.PrintSummary(summary)
}
