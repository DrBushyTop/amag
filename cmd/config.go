package cmd

import (
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "config allows you to set up default values for the tool",
}

var configSetCmd = &cobra.Command{
	Use:   "set",
	Short: "set allows you to set a default value for a given configuration key",
}

var configLoadCmd = &cobra.Command{
	Use:   "load",
	Short: "load allows you to load a configuration file",
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configLoadCmd)
}