package cmd

import (
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   `gorex`,
		Short: `Scan folder with advanced regex configurations`,
		Long:  `***`,
	}
)

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}
