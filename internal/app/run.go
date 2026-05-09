package app

import (
	"os"

	"github.com/def4alt/logstat/internal/parser"
)

type Config struct {
	FilePath string
	JSON     bool
	TopK     int
	Strict   bool
}

func Run(config Config) error {
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

		parser.ProcessLog(file, config.TopK)
	} else {
		parser.ProcessLog(os.Stdin, config.TopK)
	}

	return nil
}
