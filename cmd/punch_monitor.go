package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"connectcli/internal/punchmonitor"
	"connectcli/internal/utils"
)

var punchMonitorPeriod string

var punchMonitorCmd = &cobra.Command{
	Use:    "__punch-monitor",
	Hidden: true,
	Short:  "Internal: poll clock status and send Slack on an interval",
	RunE: func(cmd *cobra.Command, args []string) error {
		s := punchMonitorPeriod
		if s == "" {
			s = "00:10"
		}
		d, err := utils.ParseHHMMPeriod(s)
		if err != nil {
			return fmt.Errorf("invalid --period: %w", err)
		}
		return punchmonitor.RunMonitor(d)
	},
}

func init() {
	punchMonitorCmd.Flags().StringVar(&punchMonitorPeriod, "period", "", "Slack reminder interval as hh:mm (default 00:10)")
	rootCmd.AddCommand(punchMonitorCmd)
}
