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
	rootCmd.AddCommand(validateSessionCmd)
	rootCmd.AddCommand(fetchCmd)
	rootCmd.AddCommand(addshiftCmd)
}
