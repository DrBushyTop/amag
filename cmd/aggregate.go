package cmd

import (
	"github.com/spf13/cobra"
)

// aggregateCmd represents the aggregate command
var aggregateCmd = &cobra.Command{
	Use:   "aggregate",
	Short: "aggregate values from kql queries and save as custom metrics or logs",
}

func init() {
	rootCmd.AddCommand(aggregateCmd)
}