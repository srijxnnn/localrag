package main

import (
	"os"

	"github.com/srijxnnn/localrag/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
