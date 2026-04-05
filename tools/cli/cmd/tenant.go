package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/Camionerou/rag-saldivia/tools/pkg/admin"
	"github.com/spf13/cobra"
)

var tenantCmd = &cobra.Command{
	Use:   "tenant",
	Short: "Manage tenants",
}

var tenantListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all tenants",
	Run:   runTenantList,
}

var tenantStatusCmd = &cobra.Command{
	Use:   "status [slug]",
	Short: "Show detailed status of a tenant",
	Args:  cobra.ExactArgs(1),
	Run:   runTenantStatus,
}

func init() {
	tenantCmd.AddCommand(tenantListCmd)
	tenantCmd.AddCommand(tenantStatusCmd)
}

func getPlatformDBURL() string {
	url := env("POSTGRES_PLATFORM_URL", "")
	if url == "" {
		url = env("SDA_PLATFORM_DB", "")
	}
	if url == "" {
		url = "postgres://sda:sda_dev@localhost:5432/sda_platform?sslmode=disable"
	}
	return url
}

func runTenantList(cmd *cobra.Command, args []string) {
	dbURL := getPlatformDBURL()

	tenants, err := admin.TenantList(dbURL)
	if err != nil {
		exitErr("failed to list tenants: %v", err)
	}

	if len(tenants) == 0 {
		fmt.Println("No tenants found.")
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "SLUG\tNAME\tPLAN\tENABLED\tCREATED\n")
	for _, t := range tenants {
		fmt.Fprintf(w, "%s\t%s\t%s\t%v\t%s\n",
			t.Slug, t.Name, t.PlanID, t.Enabled, t.CreatedAt.Format("2006-01-02"))
	}
	w.Flush()
}

func runTenantStatus(cmd *cobra.Command, args []string) {
	dbURL := getPlatformDBURL()
	slug := args[0]

	detail, err := admin.TenantStatus(dbURL, slug)
	if err != nil {
		exitErr("failed to get tenant status: %v", err)
	}

	fmt.Printf("Tenant: %s\n", detail.Name)
	fmt.Printf("  Slug:      %s\n", detail.Slug)
	fmt.Printf("  ID:        %s\n", detail.ID)
	fmt.Printf("  Plan:      %s\n", detail.PlanID)
	fmt.Printf("  Enabled:   %v\n", detail.Enabled)
	fmt.Printf("  Modules:   %d enabled\n", detail.ModuleCount)
	fmt.Printf("  Created:   %s\n", detail.CreatedAt.Format("2006-01-02 15:04:05"))
}
