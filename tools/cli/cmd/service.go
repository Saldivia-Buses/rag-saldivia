package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/Camionerou/rag-saldivia/tools/pkg/admin"
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

var serviceLogsLines int

var serviceLogsCmd = &cobra.Command{
	Use:   "logs [service]",
	Short: "Show recent logs from a service container",
	Args:  cobra.ExactArgs(1),
	Run:   runServiceLogs,
}

func init() {
	serviceLogsCmd.Flags().IntVarP(&serviceLogsLines, "lines", "n", 50, "Number of log lines to show")
	serviceCmd.AddCommand(serviceHealthCmd)
	serviceCmd.AddCommand(serviceLogsCmd)
}

func runServiceHealth(cmd *cobra.Command, args []string) {
	baseHost := env("SDA_HOST", "localhost")
	results := admin.ServiceHealth(baseHost)

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "SERVICE\tPORT\tSTATUS\tLATENCY\n")
	for _, s := range results {
		latency := "-"
		if s.Latency > 0 {
			latency = s.Latency.Round(time.Millisecond).String()
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", s.Name, s.Port, s.Status, latency)
	}
	w.Flush()
}

func runServiceLogs(cmd *cobra.Command, args []string) {
	serviceName := args[0]

	output, err := admin.ServiceLogs(serviceName, serviceLogsLines)
	if err != nil {
		exitErr("%v", err)
	}

	fmt.Print(output)
}
