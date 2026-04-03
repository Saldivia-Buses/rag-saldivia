package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var dbCmd = &cobra.Command{
	Use:   "db",
	Short: "Database management",
}

var dbMigrateCmd = &cobra.Command{
	Use:   "migrate [tenant]",
	Short: "Run pending migrations for a tenant",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Running migrations for tenant %q...\n", args[0])
		fmt.Println("  TODO: implement migration runner via tools/pkg/admin")
	},
}

var dbBackupCmd = &cobra.Command{
	Use:   "backup [tenant]",
	Short: "Backup a tenant database",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Backing up tenant %q...\n", args[0])
		fmt.Println("  TODO: implement pg_dump wrapper via tools/pkg/admin")
	},
}

var dbSeedCmd = &cobra.Command{
	Use:   "seed [tenant] [dataset]",
	Short: "Seed a tenant database with test data",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Seeding tenant %q with dataset %q...\n", args[0], args[1])
		fmt.Println("  TODO: implement seed runner via tools/pkg/admin")
	},
}

func init() {
	dbCmd.AddCommand(dbMigrateCmd)
	dbCmd.AddCommand(dbBackupCmd)
	dbCmd.AddCommand(dbSeedCmd)
}
