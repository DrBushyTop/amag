package kql

import (
	"fmt"
	"github.com/DrBushytop/AzureMetricAggregator/pkg/auth"
	"io"
	"net/http"
)

type CustomMetricsClient struct {
	authClient *auth.Client
	httpClient *http.Client
}

func NewCustomMetricsClient(opts ...CustomMetricClientOption) (*CustomMetricsClient, error) {
	authClient, err := auth.NewAuthClient()
	if err != nil {
		return nil, err
	}

	httpClient := http.DefaultClient

	return &CustomMetricsClient{
		authClient: authClient,
		httpClient: httpClient,
	}, nil
}

type CustomMetricClientOption func(client *CustomMetricsClient) error

func WithAuthClient(authClient *auth.Client) CustomMetricClientOption {
	return func(client *CustomMetricsClient) error {
		client.authClient = authClient
		return nil
	}
}

func WithHttpClient(httpClient *http.Client) CustomMetricClientOption {
	return func(client *CustomMetricsClient) error {
		client.httpClient = httpClient
		return nil
	}
}

func (c *CustomMetricsClient) SendCustomMetrics() error {
	token, err := c.authClient.GetAccessToken([]string{"https://monitoring.azure.com/.default"})
	if err != nil {
		return fmt.Errorf("SendCustomMetrics: failed to get access token: %w", err)
	}
	request := http.Request{
		Header: map[string][]string{
			"Authorization": {fmt.Sprintf("Bearer %s", token)},
		},
	}

	res, err := c.httpClient.Do(&request)
	if err != nil {
		return fmt.Errorf("SendCustomMetrics: failed to send request: %w", err)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(res.Body)

	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("SendCustomMetrics: failed to read response body: %w", err)
	}

	// Convert body to string for logging or error message
	bodyString := string(bodyBytes)

	// Check for non-200 status code
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("SendCustomMetrics: request failed: status %d, %s, response body: %s", res.StatusCode, res.Status, bodyString)
	}

	return nil
}