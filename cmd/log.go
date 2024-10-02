package cmd

import (
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/monitor/azquery"
	"github.com/DrBushytop/amag/pkg/kql"
	"github.com/charmbracelet/log"
	"github.com/spf13/viper"
	"golang.org/x/net/context"
	"time"

	"github.com/spf13/cobra"
)

// logCmd represents the log command
var logCmd = &cobra.Command{
	Use:   "log",
	Short: "Aggregate KQL values and save as logs",
	Long: `Run a specified KQL file against an Azure Log Analytics workspace, aggregate the result,
and save it as a custom metric in Azure Monitor. This command is useful for transforming
Log Analytics query results into metrics that can be monitored and visualized over time.

Example usage:

amag aggregate metric --file "/path/to/query.kql" --metric "LatencyP90" --workspaceid "<workspace-id>" --datacollectionendpoint "<data-collection-endpoint>" --datacollectionstreamname "<data-collection-stream-name>" --datacollectionruleid "<data-collection-rule-id>"

You can set defaults using the config command or env variables.

This command requires:
- A KQL query file that defines the aggregation.
- A valid workspace ID where the query will be executed.
- Name of the metric. It will be shown in the metricName column of the Azure Log Analytics table.
- A data collection endpoint to send data to.
- A data collection stream name to send data to.
- A data collection rule ID to use.
`,
	Run: RunAggregateLog,
}

func RunAggregateLog(cmd *cobra.Command, args []string) {
	keys := viper.AllKeys()
	_ = keys
	metricName := viper.GetString(GetViperKey(cmd, KeyMetric))
	fileName := viper.GetString(GetViperKey(cmd, KeyFile))
	workspaceId := viper.GetString(GetViperKey(cmd, KeyWorkspaceID))
	dataCollectionEndpoint := viper.GetString(GetViperKey(cmd, KeyDataCollectionEndpoint))
	dataCollectionStreamName := viper.GetString(GetViperKey(cmd, KeyDataCollectionStreamName))
	dataCollectionRuleId := viper.GetString(GetViperKey(cmd, KeyDataCollectionRuleId))

	cmd.SilenceUsage = true // Avoid printing usage on error generated by functions later

	query, err := kql.ParseQuery(fileName)
	if err != nil {
		log.Error("Error parsing query from file", "file", fileName, "err", err)
		return
	}

	log.Infof("Running Query:\n%s", query)

	wsClient, err := kql.NewWorkspaceClient(workspaceId)
	if err != nil {
		log.Error("Failed to create workspace client", "err", err)
		return
	}

	res, err := wsClient.QueryWorkspaceForAggregateValue(
		context.Background(),
		azquery.Body{
			Query:    to.Ptr(query),
			Timespan: to.Ptr(azquery.NewTimeInterval(time.Now().Add(time.Duration(-24)*time.Hour), time.Now())),
		},
		nil,
	)
	if err != nil {
		log.Error("Failed to aggregate workspace", "err", err)
		return
	}

	logsClient, err := kql.NewLogsClient(dataCollectionStreamName, dataCollectionEndpoint, dataCollectionRuleId)
	if err != nil {
		log.Error("Failed to create logs client", "err", err)
		return
	}

	ag := kql.AggregateLogEntry{
		TimeGenerated: time.Now(),
		Name:          metricName,
		Value:         res,
	}

	log.Info("Sending log")
	if err := logsClient.SaveLogEntryToLogAnalytics(context.Background(), []kql.AggregateLogEntry{ag}); err != nil {
		log.Error("Failed to send log", "err", err)
		return
	}
	log.Info("Saved log", "metricName", metricName, "metricValue", res)
}

func init() {
	aggregateCmd.AddCommand(logCmd)

	err := bind(logCmd, KeyFile, "f", "", "Path to the KQL file to run")
	if err != nil {
		panic(err)
	}
	err = bind(logCmd, KeyMetric, "m", "", "Name of the custom metric to save the result into")
	if err != nil {
		panic(err)
	}
	err = bind(logCmd, KeyWorkspaceID, "w", "", "Workspace id (not the resource id) of the Log Analytics workspace to run the aggregate against")
	if err != nil {
		panic(err)
	}

	err = bind(logCmd, KeyDataCollectionEndpoint, "e", "", "The data collection endpoint to send data to")
	if err != nil {
		panic(err)
	}
	err = bind(logCmd, KeyDataCollectionStreamName, "s", "", "The data collection stream name to send data to")
	if err != nil {
		panic(err)
	}
	err = bind(logCmd, KeyDataCollectionRuleId, "r", "", "The data collection rule ID to use")
	if err != nil {
		panic(err)
	}

}