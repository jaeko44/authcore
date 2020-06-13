package app

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "authcorectl",
	Short: "Authcore CLI",
}

// Execute is the entry point of the CLI.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(migrationCmd)
	rootCmd.AddCommand(setupCmd)
	rootCmd.AddCommand(debugCmd)
}
