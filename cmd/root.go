package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "connectcli",
	Short: "A CLI tool to manage the Connecteam app",
	Long:  `ConnectCLI is a command-line interface tool for managing the Connecteam application.`,
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: true,
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if punchOutFlag {
			return runPunchOut()
		}
		if punchInFlag != "" {
			return runPunchIn()
		}
		return cmd.Help()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	return nil
}

func init() {
	rootCmd.Flags().StringVarP(&punchInFlag, "punchin", "i", "", "Punch in to a client (file path or UUID)")
	rootCmd.Flags().BoolVarP(&punchOutFlag, "punchout", "o", false, "Punch out from current shift")
	rootCmd.Flags().StringVarP(&punchNoteFlag, "note", "n", "", "Note for punch out")

	rootCmd.AddCommand(validateSessionCmd)
	rootCmd.AddCommand(fetchCmd)
	rootCmd.AddCommand(addshiftCmd)
}
