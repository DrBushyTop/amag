package cmd

import (
	"fmt"
	"github.com/spf13/pflag"
	"os"

	"github.com/charmbracelet/log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "amag",
	Short: "Amag is a tool for aggregating metrics from Kusto Query Language files as custom metrics in Azure Monitor",
	Long: `Amag is a tool for aggregating metrics from Kusto Query Language files as custom metrics in Azure Monitor. 
	It can be used to run a given KQL query and save the MetricValue property generated by the result into a given custom metric name.
	
	The tool is designed to be used in conjunction with Azure Identity, and handles authentication using DefaultAzureCredential.
	So in most cases, you'd be using this while logged in to Azure CLI or Azure Powershell.
`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(func() {
		initConfig()
		postInitCommands(rootCmd.Commands())
	})
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.amag/config.yaml)")
}

func postInitCommands(commands []*cobra.Command) {
	for _, cmd := range commands {
		presetRequiredFlags(cmd)
		if cmd.HasSubCommands() {
			postInitCommands(cmd.Commands())
		}
	}
}

func presetRequiredFlags(cmd *cobra.Command) {
	viper.BindPFlags(cmd.Flags())
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		if viper.IsSet(f.Name) && viper.GetString(f.Name) != "" {
			cmd.Flags().Set(f.Name, viper.GetString(f.Name))
		}
	})
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Println("Error finding home directory:", err)
			os.Exit(1)
		}

		configPath := home + "/.amag"

		// If the folder doesn't exist, create it
		err = os.MkdirAll(configPath, os.ModePerm)
		if err != nil {
			fmt.Printf("Error creating config folder: %v\n", err)
			return
		}

		viper.AddConfigPath(configPath)
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		log.Infof("Using config file: %s", viper.ConfigFileUsed())
	}
}