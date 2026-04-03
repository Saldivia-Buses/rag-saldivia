package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

var deployCmd = &cobra.Command{
	Use:   "deploy [service]",
	Short: "Deploy a service to production",
	Long:  "Pulls the latest Docker image and restarts the service with zero downtime.",
	Args:  cobra.ExactArgs(1),
	Run:   runDeploy,
}

var deployDryRun bool

func init() {
	deployCmd.Flags().BoolVar(&deployDryRun, "dry-run", false, "Preview deploy without executing")
}

func runDeploy(cmd *cobra.Command, args []string) {
	service := args[0]

	validServices := map[string]bool{
		"auth": true, "ws": true, "chat": true, "rag": true,
		"notification": true, "platform": true, "ingest": true, "web": true,
	}
	if !validServices[service] {
		exitErr("unknown service: %s. Valid: %s", service, strings.Join(keys(validServices), ", "))
	}

	fmt.Printf("Deploying %s...\n", service)

	if service == "web" {
		fmt.Println("  Frontend deploy is handled by Vercel/Fly.io.")
		fmt.Println("  Push to main branch to trigger deploy.")
		return
	}

	composeFile := env("SDA_COMPOSE_FILE", "deploy/docker-compose.prod.yml")
	steps := []struct {
		desc string
		cmd  string
		args []string
	}{
		{"Pulling latest image", "docker", []string{"compose", "-f", composeFile, "pull", service}},
		{"Restarting service", "docker", []string{"compose", "-f", composeFile, "up", "-d", "--no-deps", service}},
	}

	for _, step := range steps {
		fmt.Printf("  %s...\n", step.desc)
		if deployDryRun {
			fmt.Printf("    [dry-run] %s %s\n", step.cmd, strings.Join(step.args, " "))
			continue
		}

		out, err := exec.Command(step.cmd, step.args...).CombinedOutput()
		if err != nil {
			fmt.Fprintf(os.Stderr, "    FAILED: %s\n%s\n", err, string(out))
			os.Exit(1)
		}
	}

	fmt.Printf("  %s deployed successfully.\n", service)
}

func keys(m map[string]bool) []string {
	result := make([]string, 0, len(m))
	for k := range m {
		result = append(result, k)
	}
	return result
}
