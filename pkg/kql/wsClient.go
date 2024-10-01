package kql

import (
	"context"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/monitor/azquery"
	"log"
	"strconv"
)

type queryClient interface {
	QueryWorkspace(ctx context.Context, workspaceID string, body azquery.Body, options *azquery.LogsClientQueryWorkspaceOptions) (azquery.LogsClientQueryWorkspaceResponse, error)
}

type WorkspaceClient struct {
	cred        azcore.TokenCredential
	client      queryClient
	workspaceId string
}

func NewWorkspaceClient(workspaceId string, opts ...WsOption) (*WorkspaceClient, error) {
	wsc := WorkspaceClient{}

	for _, opt := range opts {
		err := opt(&wsc)
		if err != nil {
			return nil, fmt.Errorf("NewWorkspaceClient: failed to apply option: %w", err)
		}
	}

	if wsc.client == nil && wsc.cred == nil {
		cred, err := azidentity.NewDefaultAzureCredential(nil)
		if err != nil {
			return nil, fmt.Errorf("NewWorkspaceClient: failed to create default azure credential: %w", err)
		}
		wsc.cred = cred
	}

	if wsc.client == nil {
		client, err := azquery.NewLogsClient(wsc.cred, &azquery.LogsClientOptions{})
		if err != nil {
			return nil, fmt.Errorf("NewWorkspaceClient: failed to create default logs client: %w", err)
		}
		wsc.client = client
	}

	wsc.workspaceId = workspaceId

	return &wsc, nil
}

type WsOption func(client *WorkspaceClient) error

func WithQueryClient(client queryClient) WsOption {
	return func(wsc *WorkspaceClient) error {
		wsc.client = client
		return nil
	}
}

func WithCredential(cred azcore.TokenCredential) WsOption {
	return func(wsc *WorkspaceClient) error {
		wsc.cred = cred
		return nil
	}
}

// QueryWorkspaceForAggregateValue queries the workspace with the given body and options and returns the first value of the result.
// The result is expected to have ak column named 'MetricValue' and the first row of the result is returned as the single value.
// If the 'MetricValue' column is not found, an error is returned.
func (wsc *WorkspaceClient) QueryWorkspaceForAggregateValue(ctx context.Context, body azquery.Body, options *azquery.LogsClientQueryWorkspaceOptions) (float64, error) {
	result, err := wsc.client.QueryWorkspace(ctx, wsc.workspaceId, body, options)
	if err != nil {
		return 0, fmt.Errorf("QueryWorkspaceForAggregateValue: failed to query workspace: %w", err)
	}

	if result.Error != nil {
		log.Printf("QueryWorkspaceForAggregateValue: query partially failed with error: %s\n", *result.Error)
	}

	if len(result.Tables) == 0 || len(result.Tables) > 1 {
		return 0, fmt.Errorf("QueryWorkspaceForAggregateValue: unexpected number of tables found in the result. Expected 1, got %d", len(result.Tables))
	}

	if len(result.Tables[0].Columns) == 0 {
		return 0, fmt.Errorf("QueryWorkspaceForAggregateValue: no columns found in the result")
	}

	metricValueIndex := -1
	for i, col := range result.Tables[0].Columns {
		if *col.Name == "MetricValue" {
			metricValueIndex = i
			break
		}
	}
	if metricValueIndex == -1 {
		columnNames := make([]string, len(result.Tables[0].Columns))
		for i, col := range result.Tables[0].Columns {
			columnNames[i] = *col.Name
		}
		return 0, fmt.Errorf("QueryWorkspaceForAggregateValue: 'MetricValue' column not found in the result. Found columns: %v", columnNames)
	}

	metricValue := result.Tables[0].Rows[0][metricValueIndex]
	switch v := metricValue.(type) {
	case float64:
		// Metric value is already a float
		return v, nil
	case float32:
		// Convert to float64 if it's a float32
		return float64(v), nil
	case int:
		// Convert integer to float64
		return float64(v), nil
	case string:
		// Try to parse the string as a float
		value, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return 0, fmt.Errorf("QueryWorkspaceForAggregateValue: failed to parse MetricValue %v as float: %w", metricValue, err)
		}
		return value, nil
	default:
		return 0, fmt.Errorf("QueryWorkspaceForAggregateValue: unexpected MetricValue type %T, value: %v", metricValue, metricValue)
	}
}