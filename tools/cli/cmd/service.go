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

var serviceCmd = &cobra.Command{
	Use:   "service",
	Short: "Manage SDA services",
}

var serviceHealthCmd = &cobra.Command{
	Use:   "health",
	Short: "Check health of all services",
	Run:   runServiceHealth,
}

func init() {
	serviceCmd.AddCommand(serviceHealthCmd)
}

type serviceInfo struct {
	Name string
	Port string
}

var services = []serviceInfo{
	{Name: "auth", Port: "8001"},
	{Name: "ws", Port: "8002"},
	{Name: "chat", Port: "8003"},
	{Name: "rag", Port: "8004"},
	{Name: "notification", Port: "8005"},
	{Name: "platform", Port: "8006"},
	{Name: "ingest", Port: "8007"},
}

func runServiceHealth(cmd *cobra.Command, args []string) {
	baseHost := env("SDA_HOST", "localhost")
	client := &http.Client{Timeout: 3 * time.Second}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "SERVICE\tPORT\tSTATUS\tLATENCY\n")

	for _, svc := range services {
		url := fmt.Sprintf("http://%s:%s/health", baseHost, svc.Port)
		start := time.Now()
		resp, err := client.Get(url)
		latency := time.Since(start)

		if err != nil {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", svc.Name, svc.Port, "DOWN", "-")
			continue
		}
		resp.Body.Close()

		var body map[string]string
		if resp.StatusCode == http.StatusOK {
			json.NewDecoder(resp.Body).Decode(&body)
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", svc.Name, svc.Port, "UP", latency.Round(time.Millisecond))
		} else {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", svc.Name, svc.Port, fmt.Sprintf("ERR(%d)", resp.StatusCode), latency.Round(time.Millisecond))
		}
	}
	w.Flush()
}
