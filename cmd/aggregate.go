/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/monitor/azquery"
	"github.com/DrBushytop/AzureMetricAggregator/pkg/kql"
	"regexp"
	"time"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

// aggregateCmd represents the aggregate command
var aggregateCmd = &cobra.Command{
	Use:   "aggregate",
	Short: "aggregate metrics from kql files",
	Long: `Runs a given KQL query and saves the MetricValue property generated by the result into given custom metric name. For example:

amag aggregate -f query.kql -m metricName -w workspaceId -s scopeResourceId`,
	Run: RunAggregate,
}

func RunAggregate(cmd *cobra.Command, args []string) {
	metricName, _ := cmd.Flags().GetString("metric")
	fileName, _ := cmd.Flags().GetString("file")
	workspaceId, _ := cmd.Flags().GetString("workspaceid")
	scopeResourceId, _ := cmd.Flags().GetString("scoperesourceid")
	if err := validateResourceId(scopeResourceId); err != nil {
		log.Error("Error validating scopeResourceId", "err", err)
		return
	}

	cmd.SilenceUsage = true // Avoid printing usage on error generated by functions later

	query, err := kql.ParseQuery(fileName)
	if err != nil {
		log.Error("Error parsing query", "err", err)
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
		log.Error("Failed to query workspace", "err", err)
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
	}

	log.Info("Saved custom metric", "metricName", metricName, "metricValue", res, "scope", scopeResourceId)
}

func validateResourceId(resourceId string) error {
	unifiedPattern := `^/subscriptions/([a-f0-9\-]{36})/resourceGroups/([a-zA-Z0-9_\-\.]+)/providers/([a-zA-Z0-9_\-\.]+)/([a-zA-Z0-9_\-\.]+)/([a-zA-Z0-9_\-\.]+)(?:/([a-zA-Z0-9_\-\.]+)/([a-zA-Z0-9_\-\.]+))?$`
	reUnified := regexp.MustCompile(unifiedPattern)

	if !reUnified.MatchString(resourceId) {
		return fmt.Errorf("invalid resourceId format. Metrics only support resource or subresource scope")
	}
	return nil
}

func init() {
	rootCmd.AddCommand(aggregateCmd)
	aggregateCmd.Flags().StringP("file", "f", "", "Relative path to a KQL file to run")
	_ = aggregateCmd.MarkFlagRequired("file")
	aggregateCmd.Flags().StringP("metric", "m", "", "Name of the custom metric to save the result into")
	_ = aggregateCmd.MarkFlagRequired("metric")
	aggregateCmd.Flags().StringP("workspaceid", "w", "", "Workspace id (not the resource id) of the Log Analytics workspace to run the query against")
	_ = aggregateCmd.MarkFlagRequired("workspaceid")
	aggregateCmd.Flags().StringP("scoperesourceid", "s", "", "Resource id of the scope to save the custom metric to")
	_ = aggregateCmd.MarkFlagRequired("scoperesourceid")
}