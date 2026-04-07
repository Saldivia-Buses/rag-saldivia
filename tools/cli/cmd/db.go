package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Camionerou/rag-saldivia/tools/pkg/admin"
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

var dbQueryCmd = &cobra.Command{
	Use:   "query [tenant] [sql]",
	Short: "Execute a read-only SELECT query against a tenant database",
	Long: `Execute a read-only SQL query against a tenant's database.
Only SELECT queries are allowed. Results are limited to 100 rows.

  sda db query dev "SELECT id, email FROM users LIMIT 5"`,
	Args: cobra.ExactArgs(2),
	Run:  runDBQuery,
}

func init() {
	dbCmd.AddCommand(dbMigrateCmd)
	dbCmd.AddCommand(dbSeedCmd)
	dbCmd.AddCommand(dbBackupCmd)
	dbCmd.AddCommand(dbQueryCmd)
}

func runDBQuery(cmd *cobra.Command, args []string) {
	tenant := args[0]
	query := strings.TrimSpace(args[1])

	platformURL := env("POSTGRES_PLATFORM_URL", "postgres://sda:sda_dev@localhost:5432/sda_platform?sslmode=disable")
	rows, err := admin.DBQuery(platformURL, tenant, query)
	if err != nil {
		exitErr("%v", err)
	}
	if len(rows) == 0 {
		fmt.Println("No rows returned.")
		return
	}
	out, _ := json.MarshalIndent(rows, "", "  ")
	fmt.Printf("%d rows:\n%s\n", len(rows), string(out))
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
