package kql

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/DrBushytop/amag/pkg/auth"
	"golang.org/x/net/context"
	"io"
	"net/http"
	"strings"
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

func (c *CustomMetricsClient) SendCustomMetrics(ctx context.Context, scopeResourceId string, location string, body CustomMetricBody) error {
	token, err := c.authClient.GetAccessToken([]string{"https://monitoring.azure.com/"})
	if err != nil {
		return fmt.Errorf("SendCustomMetrics: failed to get access token: %w", err)
	}

	// Remove the leading slash from the resourceId as expected by the API
	scopeResourceId, _ = strings.CutPrefix(scopeResourceId, "/")
	uriString := fmt.Sprintf("https://%s.monitoring.azure.com/%s/metrics", location, scopeResourceId)
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("SendCustomMetrics: failed to marshal body: %w", err)
	}

	request, err := http.NewRequestWithContext(ctx, "POST", uriString, bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("SendCustomMetrics: failed to create request: %w", err)
	}
	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	request.Header.Add("Content-Type", "application/json")

	res, err := c.httpClient.Do(request)
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