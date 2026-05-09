// Command logstat reads log files and outputs summary statistics
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/def4alt/logstat/internal/app"
)

var (
	fileFlag = flag.String("file", "", "Path to the log file")
	jsonFlag = flag.Bool("json", false, "Output results in JSON format")
	topNFlag = flag.Int("top", 10, "Number of top entries to display")
	strict   = flag.Bool("strict", false, "Enable strict parsing of log entries")
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `Usage: logstat [options]

Analyzes log files and outputs summary statistics.

Options:
	--file <path>   Path to the log file (reads from stdin if omitted)
	--json          Output results in JSON format
	--top <n>       Number of top entries to display (default: 10)
	--strict        Abort on first malformed line instead of skipping

Examples:
	logstat --file access.log
	cat access.log | logstat
	logstat --file access.log --json --top 20
`)
	}

	flag.Parse()

	config := app.Config{
		FilePath: *fileFlag,
		JSON:     *jsonFlag,
		TopK:     *topNFlag,
		Strict:   *strict,
	}

	if err := app.Run(config); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	}
}
