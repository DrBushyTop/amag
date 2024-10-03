package kql

import (
	"context"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/monitor/azquery"
	"log"
	"strconv"
	"time"
)

type LogLine struct {
	TimeGenerated time.Time `json:"TimeGenerated"`
	MetricValue   float64   `json:"MetricValue"`
}

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
// The result is expected to have columns named 'TimeGenerated' and 'MetricValue'. A slice of LogLine is returned, one for each row in the result.
// If the columns are not found, an error is returned.
func (wsc *WorkspaceClient) QueryWorkspaceForAggregateValue(ctx context.Context, body azquery.Body, options *azquery.LogsClientQueryWorkspaceOptions) ([]LogLine, error) {
	result, err := wsc.client.QueryWorkspace(ctx, wsc.workspaceId, body, options)
	if err != nil {
		return []LogLine{}, fmt.Errorf("QueryWorkspaceForAggregateValue: failed to query workspace: %w", err)
	}

	if result.Error != nil {
		log.Printf("QueryWorkspaceForAggregateValue: query partially failed with error: %s\n", *result.Error)
	}

	if len(result.Tables) == 0 || len(result.Tables) > 1 {
		return []LogLine{}, fmt.Errorf("QueryWorkspaceForAggregateValue: unexpected number of tables found in the result. Expected 1, got %d", len(result.Tables))
	}

	if len(result.Tables[0].Columns) == 0 {
		return []LogLine{}, fmt.Errorf("QueryWorkspaceForAggregateValue: no columns found in the result")
	}

	metricValueIndex := -1
	timeGeneratedIndex := -1
	for i, col := range result.Tables[0].Columns {
		if *col.Name == "MetricValue" {
			metricValueIndex = i
			continue
		}
		if *col.Name == "TimeGenerated" {
			timeGeneratedIndex = i
			continue
		}
	}
	if metricValueIndex == -1 || timeGeneratedIndex == -1 {
		columnNames := make([]string, len(result.Tables[0].Columns))
		for i, col := range result.Tables[0].Columns {
			columnNames[i] = *col.Name
		}
		if metricValueIndex == -1 {
			return []LogLine{}, fmt.Errorf("QueryWorkspaceForAggregateValue: 'MetricValue' column not found in the result. Found columns: %v", columnNames)
		} else {
			return []LogLine{}, fmt.Errorf("QueryWorkspaceForAggregateValue: 'TimeGenerated' column not found in the result. Found columns: %v", columnNames)
		}
	}

	layout := "2006-01-02T15:04:05Z"
	res := make([]LogLine, len(result.Tables[0].Rows))
	for _, row := range result.Tables[0].Rows {
		timeGenerated := row[timeGeneratedIndex].(string)
		parsedTime, err := time.Parse(layout, timeGenerated)
		if err != nil {
			return []LogLine{}, fmt.Errorf("QueryWorkspaceForAggregateValue: failed to parse TimeGenerated %v as time: %w", timeGenerated, err)
		}
		metricValue := row[metricValueIndex]
		switch v := metricValue.(type) {
		case float64:
			// Metric value is already a float
			res = append(res, LogLine{parsedTime, v})
		case float32:
			// Convert to float64 if it's a float32
			res = append(res, LogLine{parsedTime, float64(v)})
		case int:
			// Convert integer to float64
			res = append(res, LogLine{parsedTime, float64(v)})
		case string:
			// Try to parse the string as a float
			value, err := strconv.ParseFloat(v, 64)
			if err != nil {
				return []LogLine{}, fmt.Errorf("QueryWorkspaceForAggregateValue: failed to parse MetricValue %v as float: %w", metricValue, err)
			}
			res = append(res, LogLine{parsedTime, value})
		default:
			return []LogLine{}, fmt.Errorf("QueryWorkspaceForAggregateValue: unexpected MetricValue type %T, value: %v", metricValue, metricValue)
		}
	}
	return res, nil
}