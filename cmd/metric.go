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

var metricCmd = &cobra.Command{
	Use:   "metric",
	Short: "Run a KQL query and save the result as a custom Azure Monitor metric",
	Long: `Run a specified KQL file against an Azure Log Analytics workspace, aggregate the result,
and save it as a custom metric in Azure Monitor. This command is useful for transforming
Log Analytics query results into metrics that can be monitored and visualized over time.

Example usage:

amag aggregate metric --file /path/to/query.kql --metric LatencyP90 --workspaceid <workspace-id> --scoperesourceid <scope-resource-id>

This command requires:
- A KQL query file that defines the aggregation.
- A valid workspace ID where the query will be executed.
- A scope resource ID where the custom metric will be saved. This can be a resource or subresource ID.`,
	Run: RunAggregateMetric,
}

func RunAggregateMetric(cmd *cobra.Command, args []string) {
	metricName := viper.GetString(GetViperKey(cmd, KeyMetric))
	fileName := viper.GetString(GetViperKey(cmd, KeyFile))
	workspaceId := viper.GetString(GetViperKey(cmd, KeyWorkspaceID))
	scopeResourceId := viper.GetString(GetViperKey(cmd, KeyScopeResourceID))
	if err := validateResourceId(scopeResourceId); err != nil {
		log.Error("Error validating scopeResourceId", "err", err)
		return
	}

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

	body := kql.NewCustomMetricsBody(metricName, res)

	cmClient, err := kql.NewCustomMetricsClient()
	if err != nil {
		log.Error("Failed to create custom metrics client", "err", err)
		return
	}
	log.Info("Sending custom metric")
	if err := cmClient.SendCustomMetrics(context.Background(), scopeResourceId, "westeurope", body); err != nil {
		log.Error("Failed to send custom metrics", "err", err)
		return
	}

	log.Info("Saved custom metric", "metricName", metricName, "metricValue", res, "scope", scopeResourceId)
}

func init() {
	aggregateCmd.AddCommand(metricCmd)

	err := bind(metricCmd, KeyFile, "f", "", "Path to the KQL file to run")
	if err != nil {
		panic(err)
	}
	err = bind(metricCmd, KeyMetric, "m", "", "Name of the custom metric to save the result into")
	if err != nil {
		panic(err)
	}
	err = bind(metricCmd, KeyWorkspaceID, "w", "", "Workspace id (not the resource id) of the Log Analytics workspace to run the aggregate against")
	if err != nil {
		panic(err)
	}
	err = bind(metricCmd, KeyScopeResourceID, "s", "", "Resource id of the scope to save the custom metric to")
	if err != nil {
		panic(err)
	}
}