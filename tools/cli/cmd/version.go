package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Set at build time via -ldflags.
var (
	Version   = "dev"
	GitCommit = "none"
	BuildDate = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("sda %s (commit: %s, built: %s)\n", Version, GitCommit, BuildDate)
	},
}
