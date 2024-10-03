package kql

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/monitor/ingestion/azlogs"
)

type ingestClient interface {
	Upload(ctx context.Context, ruleID string, streamName string, logs []byte, options *azlogs.UploadOptions) (azlogs.UploadResponse, error)
}

type LogsClient struct {
	cred         azcore.TokenCredential
	client       ingestClient
	dcStreamName string
	dcEndpoint   string
	dcRuleId     string
}

func NewLogsClient(dcStreamName, dcEndpoint, dcRuleId string, opts ...LogsClientOption) (*LogsClient, error) {
	logsClient := LogsClient{}

	for _, opt := range opts {
		err := opt(&logsClient)
		if err != nil {
			return nil, fmt.Errorf("NewLogsClient: failed to apply option: %w", err)
		}
	}

	if logsClient.client == nil && logsClient.cred == nil {
		cred, err := azidentity.NewDefaultAzureCredential(nil)
		if err != nil {
			return nil, fmt.Errorf("NewLogsClient: failed to create default azure credential: %w", err)
		}
		logsClient.cred = cred
	}

	if logsClient.client == nil {
		azClient, err := azlogs.NewClient(dcEndpoint, logsClient.cred, nil)
		if err != nil {
			return nil, fmt.Errorf("unable to create client: %w", err)
		}
		logsClient.client = azClient
	}

	logsClient.dcStreamName = dcStreamName
	logsClient.dcEndpoint = dcEndpoint
	logsClient.dcRuleId = dcRuleId

	return &logsClient, nil
}

type LogsClientOption func(client *LogsClient) error

func WithIngestClient(client ingestClient) LogsClientOption {
	return func(logsClient *LogsClient) error {
		logsClient.client = client
		return nil
	}
}

func WithIngestCredential(cred azcore.TokenCredential) LogsClientOption {
	return func(logsClient *LogsClient) error {
		logsClient.cred = cred
		return nil
	}
}

func (lc *LogsClient) SaveLogEntryToLogAnalytics(ctx context.Context, entry []AggregateLogEntry) error {

	logs, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("unable to marshal log entry: %w", err)
	}

	_, err = lc.client.Upload(ctx, lc.dcRuleId, lc.dcStreamName, logs, nil)
	if err != nil {
		return fmt.Errorf("unable to upload logs: %w", err)
	}

	return nil
}