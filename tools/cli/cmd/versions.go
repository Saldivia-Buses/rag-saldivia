package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var versionsCmd = &cobra.Command{
	Use:   "versions",
	Short: "Show running service versions vs current git SHA",
	Run:   runVersions,
}

type buildInfo struct {
	Service   string `json:"service"`
	Version   string `json:"version"`
	GitSHA    string `json:"git_sha"`
	BuildTime string `json:"build_time"`
	GoVersion string `json:"go_version"`
}

func runVersions(cmd *cobra.Command, args []string) {
	out, err := exec.Command("git", "rev-parse", "--short", "HEAD").Output()
	currentSHA := strings.TrimSpace(string(out))
	if err != nil || currentSHA == "" {
		currentSHA = "unknown"
	}

	fmt.Printf("Current HEAD: %s\n\n", currentSHA)
	fmt.Printf("%-20s %-10s %-10s %-22s %s\n", "SERVICE", "VERSION", "GIT SHA", "BUILD TIME", "STATUS")
	fmt.Println(strings.Repeat("─", 80))

	services := []struct {
		name string
		port int
	}{
		{"auth", 8001}, {"ws", 8002}, {"chat", 8003}, {"agent", 8004},
		{"notification", 8005}, {"platform", 8006}, {"ingest", 8007},
		{"feedback", 8008}, {"traces", 8009}, {"search", 8010},
		{"bigbrother", 8012}, {"erp", 8013},
	}

	client := &http.Client{Timeout: 2 * time.Second}

	for _, svc := range services {
		url := fmt.Sprintf("http://localhost:%d/v1/info", svc.port)
		resp, err := client.Get(url)
		if err != nil {
			fmt.Printf("%-20s %-10s %-10s %-22s \033[31mDOWN\033[0m\n",
				svc.name, "-", "-", "-")
			continue
		}

		var info buildInfo
		if decErr := json.NewDecoder(resp.Body).Decode(&info); decErr != nil {
			_ = resp.Body.Close()
			fmt.Printf("%-20s %-10s %-10s %-22s \033[33mBAD RESPONSE\033[0m\n",
				svc.name, "-", "-", "-")
			continue
		}
		_ = resp.Body.Close()

		status := "\033[32mMATCH\033[0m"
		if info.GitSHA != currentSHA {
			if info.GitSHA == "unknown" || info.GitSHA == "" {
				status = "\033[33mNO INFO\033[0m"
			} else {
				status = "\033[31mSTALE\033[0m"
			}
		}

		fmt.Printf("%-20s %-10s %-10s %-22s %s\n",
			svc.name, info.Version, info.GitSHA, info.BuildTime, status)
	}
}
