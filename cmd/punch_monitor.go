package cmd

import (
	"github.com/spf13/cobra"

	"connectcli/internal/punchmonitor"
)

var punchMonitorCmd = &cobra.Command{
	Use:    "__punch-monitor",
	Hidden: true,
	Short:  "Internal: poll clock status and send Slack every 10 minutes",
	RunE: func(cmd *cobra.Command, args []string) error {
		return punchmonitor.RunMonitor()
	},
}

func init() {
	rootCmd.AddCommand(punchMonitorCmd)
}
