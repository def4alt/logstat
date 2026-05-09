package app

import (
	"fmt"
	"os"
	"text/tabwriter"

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

	totalEntries := stats.TotalEntries(entries)
	totalBytes := stats.TotalBytes(entries)
	uniqueHosts := stats.UniqueHosts(entries)
	statusCodeCounts := stats.StatusCodeCounts(entries)
	methodCounts := stats.MethodCounts(entries)
	topkHosts := stats.TopKHosts(entries, config.TopK)
	topKPaths := stats.TopKPaths(entries, config.TopK)
	p50Bytes := stats.P50Bytes(entries)
	p90Bytes := stats.P90Bytes(entries)
	p95Bytes := stats.P95Bytes(entries)
	p99Bytes := stats.P99Bytes(entries)

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	fmt.Fprintf(w, "Total Entries:\t%d\n", totalEntries)
	fmt.Fprintf(w, "Total Bytes:\t%d\n", totalBytes)
	fmt.Fprintf(w, "Unique Hosts:\t%d\n", uniqueHosts)

	fmt.Fprintln(w, "\nStatus Code Counts:")
	for code, count := range statusCodeCounts {
		fmt.Fprintf(w, "%s:\t%d\n", code, count)
	}

	fmt.Fprintln(w, "\nMethod Counts:")
	for method, count := range methodCounts {
		fmt.Fprintf(w, "%s:\t%d\n", method, count)
	}

	fmt.Fprintln(w, "\nTop Hosts:")
	for _, kv := range topkHosts {
		fmt.Fprintf(w, "%s:\t%d\n", kv.Key, kv.Value)
	}

	fmt.Fprintln(w, "\nTop Paths:")
	for _, kv := range topKPaths {
		fmt.Fprintf(w, "%s:\t%d\n", kv.Key, kv.Value)
	}

	fmt.Fprintln(w, "\nBytes Percentiles:")
	fmt.Fprintf(w, "P50:\t%d\n", p50Bytes)
	fmt.Fprintf(w, "P90:\t%d\n", p90Bytes)
	fmt.Fprintf(w, "P95:\t%d\n", p95Bytes)
	fmt.Fprintf(w, "P99:\t%d\n", p99Bytes)

	if err := w.Flush(); err != nil {
		return err
	}

	return nil
}
