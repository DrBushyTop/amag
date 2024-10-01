package cmd

import (
	"errors"
	"fmt"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "config allows you to set up default values for the tool",
}

var configSetCmd = &cobra.Command{
	Use:   "set [key] [value]",
	Short: "Set a default value for a configuration key",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]
		value := args[1]

		viper.Set(key, value)

		// Write the configuration to the file
		err := viper.WriteConfig()
		if err != nil {
			// If the config file doesn't exist, create it
			var configFileNotFoundError viper.ConfigFileNotFoundError
			if errors.As(err, &configFileNotFoundError) {
				err = viper.SafeWriteConfig()
			}
			if err != nil {
				fmt.Println("Error writing config file:", err)
			} else {
				fmt.Println("Configuration saved.")
			}
		} else {
			fmt.Println("Configuration saved.")
		}
	},
}

var configLoadCmd = &cobra.Command{
	Use:   "load [path]",
	Short: "Load configuration from a file. Must be in yaml format.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		configPath := args[0]

		viper.SetConfigFile(configPath)

		viper.SetConfigType("yaml")
		if err := viper.ReadInConfig(); err != nil {
			log.Error("Error reading config file:", err)
		} else {
			log.Info("Configuration loaded from", viper.ConfigFileUsed())
		}
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configLoadCmd)
}