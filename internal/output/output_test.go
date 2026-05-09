package output

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/def4alt/logstat/internal/stats"
	"github.com/def4alt/logstat/internal/types"
)

func captureStdout(fn func() error) (string, error) {
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		return "", err
	}
	os.Stdout = w
	defer func() { os.Stdout = old }()

	err = fn()
	w.Close()

	var buf bytes.Buffer
	_, copyErr := buf.ReadFrom(r)
	if err != nil {
		return "", err
	}
	if copyErr != nil {
		return "", copyErr
	}
	return buf.String(), nil
}

func emptySummary() stats.Summary {
	return stats.Summary{
		StatusCodeCounts: map[string]int{},
		MethodCounts:     map[string]int{},
		TopKHosts:        []types.KV[int]{},
		TopKPaths:        []types.KV[int]{},
	}
}

func TestPrintSummary(t *testing.T) {
	summary := emptySummary()
	summary.TotalEntries = 1500
	summary.TotalBytes = 5000000
	summary.UniqueHosts = 42

	out, err := captureStdout(func() error {
		return PrintSummary(summary)
	})
	if err != nil {
		t.Fatalf("PrintSummary error: %v", err)
	}

	checks := []struct {
		name string
		frag string
	}{
		{"total entries", "Total Entries:"},
		{"total entries value", "1500"},
		{"total bytes", "Total Bytes:"},
		{"unique hosts", "Unique Hosts:"},
		{"unique", "42"},
	}

	for _, c := range checks {
		t.Run(c.name, func(t *testing.T) {
			if !strings.Contains(out, c.frag) {
				t.Errorf("output missing %q\n%s", c.frag, out)
			}
		})
	}
}

func TestPrintSummaryContainsSections(t *testing.T) {
	summary := emptySummary()
	summary.TotalEntries = 100
	summary.StatusCodeCounts = map[string]int{"200": 80}
	summary.TopKHosts = []types.KV[int]{
		{Key: "host-a", Value: 50},
	}

	out, err := captureStdout(func() error {
		return PrintSummary(summary)
	})
	if err != nil {
		t.Fatalf("PrintSummary error: %v", err)
	}

	sections := []string{
		"Total Entries:",
		"Status Code Counts:",
		"200:",
		"Top Hosts:",
		"host-a:",
		"Bytes Percentiles:",
		"P50:",
	}
	for _, s := range sections {
		if !strings.Contains(out, s) {
			t.Errorf("output missing section %q\n%s", s, out)
		}
	}
}

func TestPrintSummarySkipped(t *testing.T) {
	summary := emptySummary()
	summary.SkippedEntries = 31

	out, err := captureStdout(func() error {
		return PrintSummary(summary)
	})
	if err != nil {
		t.Fatalf("PrintSummary error: %v", err)
	}

	if !strings.Contains(out, "Malformed lines skipped: 31") {
		t.Errorf("output missing skipped message\n%s", out)
	}
}

func TestPrintSummaryNoSkipped(t *testing.T) {
	summary := emptySummary()
	summary.TotalEntries = 10

	out, err := captureStdout(func() error {
		return PrintSummary(summary)
	})
	if err != nil {
		t.Fatalf("PrintSummary error: %v", err)
	}

	if strings.Contains(out, "Malformed lines skipped") {
		t.Errorf("output should not mention skipped when count is 0\n%s", out)
	}
}

func TestPrintJSON(t *testing.T) {
	summary := emptySummary()
	summary.TotalEntries = 100
	summary.TotalBytes = 5000
	summary.UniqueHosts = 5
	summary.StatusCodeCounts = map[string]int{"200": 90, "404": 10}
	summary.TopKHosts = []types.KV[int]{{Key: "h1", Value: 60}, {Key: "h2", Value: 40}}

	out, err := captureStdout(func() error {
		return PrintJSON(summary)
	})
	if err != nil {
		t.Fatalf("PrintJSON error: %v", err)
	}

	if !json.Valid([]byte(out)) {
		t.Errorf("output is not valid JSON:\n%s", out)
	}

	var decoded stats.Summary
	if err := json.Unmarshal([]byte(out), &decoded); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v\n%s", err, out)
	}

	if decoded.TotalEntries != 100 {
		t.Errorf("TotalEntries = %d, want 100", decoded.TotalEntries)
	}
	if decoded.TotalBytes != 5000 {
		t.Errorf("TotalBytes = %d, want 5000", decoded.TotalBytes)
	}
	if decoded.UniqueHosts != 5 {
		t.Errorf("UniqueHosts = %d, want 5", decoded.UniqueHosts)
	}
	if len(decoded.StatusCodeCounts) != 2 {
		t.Errorf("len(StatusCodeCounts) = %d, want 2", len(decoded.StatusCodeCounts))
	}
}

func TestPrintJSONPretty(t *testing.T) {
	summary := emptySummary()
	summary.TotalEntries = 1

	out, err := captureStdout(func() error {
		return PrintJSON(summary)
	})
	if err != nil {
		t.Fatalf("PrintJSON error: %v", err)
	}

	if !strings.Contains(out, "\n") {
		t.Errorf("expected pretty-printed (multiline) JSON, got:\n%s", out)
	}

	if !strings.Contains(out, "  ") {
		t.Errorf("expected indented JSON, got:\n%s", out)
	}
}

func TestPrintJSONEmptySummary(t *testing.T) {
	summary := emptySummary()

	out, err := captureStdout(func() error {
		return PrintJSON(summary)
	})
	if err != nil {
		t.Fatalf("PrintJSON error: %v", err)
	}

	if !json.Valid([]byte(out)) {
		t.Errorf("output is not valid JSON:\n%s", out)
	}
}
