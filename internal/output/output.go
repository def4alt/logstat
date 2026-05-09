package output

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/def4alt/logstat/internal/stats"
)

func PrintSummary(summary stats.Summary) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	if summary.SkippedEntries > 0 {
		if _, err := fmt.Fprintf(w, "Malformed lines skipped: %d\n", summary.SkippedEntries); err != nil {
			return err
		}
	}

	lines := []string{
		fmt.Sprintf("Total Entries:\t%d", summary.TotalEntries),
		fmt.Sprintf("Total Bytes:\t%d", summary.TotalBytes),
		fmt.Sprintf("Unique Hosts:\t%d", summary.UniqueHosts),
	}

	lines = append(lines, "", "Status Code Counts:")
	for code, count := range summary.StatusCodeCounts {
		lines = append(lines, fmt.Sprintf("%s:\t%d", code, count))
	}

	lines = append(lines, "", "Method Counts:")
	for method, count := range summary.MethodCounts {
		lines = append(lines, fmt.Sprintf("%s:\t%d", method, count))
	}

	lines = append(lines, "", "Top Hosts:")
	for _, kv := range summary.TopKHosts {
		lines = append(lines, fmt.Sprintf("%s:\t%d", kv.Key, kv.Value))
	}

	lines = append(lines, "", "Top Paths:")
	for _, kv := range summary.TopKPaths {
		lines = append(lines, fmt.Sprintf("%s:\t%d", kv.Key, kv.Value))
	}

	lines = append(lines, "", "Bytes Percentiles:")
	lines = append(lines, fmt.Sprintf("P50:\t%d", summary.P50Bytes))
	lines = append(lines, fmt.Sprintf("P90:\t%d", summary.P90Bytes))
	lines = append(lines, fmt.Sprintf("P95:\t%d", summary.P95Bytes))
	lines = append(lines, fmt.Sprintf("P99:\t%d", summary.P99Bytes))

	for _, line := range lines {
		if _, err := fmt.Fprintln(w, line); err != nil {
			return err
		}
	}

	if err := w.Flush(); err != nil {
		return err
	}

	return nil
}

func PrintJSON(summary stats.Summary) error {
	data, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return err
	}

	_, err = os.Stdout.Write(data)

	return err
}
