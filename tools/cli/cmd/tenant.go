package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"text/tabwriter"
	"time"

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

func init() {
	tenantCmd.AddCommand(tenantListCmd)
}

type tenant struct {
	ID        string `json:"id"`
	Slug      string `json:"slug"`
	Name      string `json:"name"`
	Enabled   bool   `json:"enabled"`
	CreatedAt string `json:"created_at"`
}

func runTenantList(cmd *cobra.Command, args []string) {
	baseHost := env("SDA_HOST", "localhost")
	token := env("SDA_TOKEN", "")
	if token == "" {
		exitErr("SDA_TOKEN environment variable is required")
	}

	client := &http.Client{Timeout: 10 * time.Second}
	url := fmt.Sprintf("http://%s:8006/v1/platform/tenants", baseHost)

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		exitErr("failed to connect to platform service: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		exitErr("platform service returned %d", resp.StatusCode)
	}

	var tenants []tenant
	if err := json.NewDecoder(resp.Body).Decode(&tenants); err != nil {
		exitErr("failed to parse response: %v", err)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "SLUG\tNAME\tENABLED\tCREATED\n")
	for _, t := range tenants {
		created := ""
		if ts, err := time.Parse(time.RFC3339, t.CreatedAt); err == nil {
			created = ts.Format("2006-01-02")
		}
		fmt.Fprintf(w, "%s\t%s\t%v\t%s\n", t.Slug, t.Name, t.Enabled, created)
	}
	w.Flush()
}
