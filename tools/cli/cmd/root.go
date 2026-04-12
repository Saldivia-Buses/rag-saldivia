package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "sda",
	Short: "SDA Framework CLI",
	Long:  "Control and manage SDA Framework — microservices, tenants, deploys, and more.",
}

// Execute is the entry point for the CLI.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(versionsCmd)
	rootCmd.AddCommand(serviceCmd)
	rootCmd.AddCommand(tenantCmd)
	rootCmd.AddCommand(deployCmd)
	rootCmd.AddCommand(dbCmd)
}

// env reads an environment variable with a fallback default.
func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// exitErr prints an error and exits.
func exitErr(msg string, args ...any) {
	fmt.Fprintf(os.Stderr, "error: "+msg+"\n", args...)
	os.Exit(1)
}
