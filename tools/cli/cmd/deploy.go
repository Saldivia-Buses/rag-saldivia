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
	Short: "Deploy services to production",
	Long: `Deploy the full stack or a single service.

  sda deploy           — deploy all services (docker compose up -d --build)
  sda deploy auth      — deploy only the auth service`,
	Args: cobra.MaximumNArgs(1),
	Run:  runDeploy,
}

var deployDryRun bool

func init() {
	deployCmd.Flags().BoolVar(&deployDryRun, "dry-run", false, "Preview deploy without executing")
}

var validServices = map[string]bool{
	"auth": true, "ws": true, "chat": true, "rag": true,
	"notification": true, "platform": true, "ingest": true,
	"feedback": true, "web": true,
}

func runDeploy(cmd *cobra.Command, args []string) {
	composeFile := env("SDA_COMPOSE_FILE", "deploy/docker-compose.prod.yml")

	// Full-stack deploy: sda deploy
	if len(args) == 0 {
		fmt.Println("Deploying full stack...")
		runStep("Building and starting all services",
			"docker", []string{"compose", "-f", composeFile, "up", "-d", "--build"},
		)
		return
	}

	// Single service deploy: sda deploy {service}
	service := args[0]
	if !validServices[service] {
		exitErr("unknown service: %s\nValid services: %s", service, strings.Join(keys(validServices), ", "))
	}

	fmt.Printf("Deploying %s...\n", service)

	if service == "web" {
		fmt.Println("  Frontend deploy is handled by Vercel/Fly.io.")
		fmt.Println("  Push to main branch to trigger deploy.")
		return
	}

	steps := []struct {
		desc string
		cmd  string
		args []string
	}{
		{"Pulling latest image", "docker", []string{"compose", "-f", composeFile, "pull", service}},
		{"Restarting service", "docker", []string{"compose", "-f", composeFile, "up", "-d", "--no-deps", service}},
	}

	for _, step := range steps {
		runStep(step.desc, step.cmd, step.args)
	}

	fmt.Printf("  %s deployed successfully.\n", service)
}

func runStep(desc, command string, args []string) {
	fmt.Printf("  %s...\n", desc)
	if deployDryRun {
		fmt.Printf("    [dry-run] %s %s\n", command, strings.Join(args, " "))
		return
	}

	c := exec.Command(command, args...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	if err := c.Run(); err != nil {
		exitErr("%s failed: %v", desc, err)
	}
}

func keys(m map[string]bool) []string {
	result := make([]string, 0, len(m))
	for k := range m {
		result = append(result, k)
	}
	return result
}
