package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

var dbCmd = &cobra.Command{
	Use:   "db",
	Short: "Database management",
}

var dbMigrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Run pending migrations for platform and tenant databases",
	Run:   runDBMigrate,
}

var dbSeedCmd = &cobra.Command{
	Use:   "seed",
	Short: "Seed databases with dev data",
	Run:   runDBSeed,
}

var dbBackupCmd = &cobra.Command{
	Use:   "backup [tenant]",
	Short: "Backup a tenant database",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Backing up tenant %q...\n", args[0])
		fmt.Println("  TODO: implement pg_dump wrapper")
	},
}

func init() {
	dbCmd.AddCommand(dbMigrateCmd)
	dbCmd.AddCommand(dbSeedCmd)
	dbCmd.AddCommand(dbBackupCmd)
}

// repoRoot finds the repository root by looking for go.work.
func repoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.work")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("could not find repo root (no go.work found)")
		}
		dir = parent
	}
}

func runDBMigrate(cmd *cobra.Command, args []string) {
	root, err := repoRoot()
	if err != nil {
		exitErr("cannot locate repo root: %v", err)
	}

	script := filepath.Join(root, "deploy", "scripts", "migrate.sh")
	if _, err := os.Stat(script); err != nil {
		exitErr("migration script not found at %s", script)
	}

	fmt.Println("Running migrations...")
	c := exec.Command("bash", script)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	// Forward DB URLs from env if set.
	c.Env = os.Environ()

	if err := c.Run(); err != nil {
		exitErr("migration failed: %v", err)
	}
}

func runDBSeed(cmd *cobra.Command, args []string) {
	root, err := repoRoot()
	if err != nil {
		exitErr("cannot locate repo root: %v", err)
	}

	script := filepath.Join(root, "deploy", "scripts", "seed.sh")
	if _, err := os.Stat(script); err != nil {
		exitErr("seed script not found at %s", script)
	}

	fmt.Println("Running seed...")
	c := exec.Command("bash", script)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Env = os.Environ()

	if err := c.Run(); err != nil {
		exitErr("seed failed: %v", err)
	}
}
