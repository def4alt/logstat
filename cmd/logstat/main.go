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
	flag.Parse()

	config := app.Config{
		FilePath: *fileFlag,
		JSON:     *jsonFlag,
		TopN:     *topNFlag,
		Strict:   *strict,
	}

	if err := app.Run(config); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	}
}
