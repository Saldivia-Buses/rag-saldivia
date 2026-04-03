package main

import (
	"os"

	"github.com/Camionerou/rag-saldivia/tools/cli/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
