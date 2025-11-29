package g5

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client handles 5G network API communications
type Client struct {
	baseURL    string
	httpClient *http.Client
	config     *Config
	timeout    time.Duration
}

// NewClient creates a new 5G API client
func NewClient(baseURL string, cfg *Config) *Client {
	if baseURL == "" {
		baseURL = "https://api.5g.vtp.local"
	}

	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		config:  cfg,
		timeout: 10 * time.Second,
	}
}

// GetNetworkStatus retrieves current network status from 5G API
func (c *Client) GetNetworkStatus(ctx context.Context) (*Network5GStatus, error) {
	url := fmt.Sprintf("%s/api/v1/network/status", c.baseURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var network Network5GStatus
	if err := json.NewDecoder(resp.Body).Decode(&network); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &network, nil
}

// MeasureLatency measures network latency via API
func (c *Client) MeasureLatency(ctx context.Context) (int, error) {
	url := fmt.Sprintf("%s/api/v1/network/latency", c.baseURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}

	start := time.Now()
	resp, err := c.httpClient.Do(req)
	duration := time.Since(start)

	if err != nil {
		return 0, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	return int(duration.Milliseconds()), nil
}

// MeasureBandwidth measures available bandwidth
func (c *Client) MeasureBandwidth(ctx context.Context) (int, error) {
	url := fmt.Sprintf("%s/api/v1/network/bandwidth", c.baseURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var result map[string]int
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, fmt.Errorf("failed to decode response: %w", err)
	}

	if bandwidth, ok := result["bandwidth"]; ok {
		return bandwidth, nil
	}

	return 0, errors.New("bandwidth not found in response")
}

// GetMetrics retrieves network metrics
func (c *Client) GetMetrics(ctx context.Context, sessionID string) (*NetworkMetrics, error) {
	if sessionID == "" {
		return nil, errors.New("sessionID is required")
	}

	url := fmt.Sprintf("%s/api/v1/metrics/network?sessionId=%s", c.baseURL, sessionID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var metrics NetworkMetrics
	if err := json.NewDecoder(resp.Body).Decode(&metrics); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &metrics, nil
}

// ReportMetrics sends network metrics to the API
func (c *Client) ReportMetrics(ctx context.Context, metrics *NetworkMetrics) error {
	if metrics == nil {
		return errors.New("metrics cannot be nil")
	}

	url := fmt.Sprintf("%s/api/v1/metrics/report", c.baseURL)

	payload, err := json.Marshal(metrics)
	if err != nil {
		return fmt.Errorf("failed to marshal metrics: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Body = io.NopCloser(io.Reader(bytes.NewReader(payload)))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// GetEdgeNodes retrieves available edge nodes
func (c *Client) GetEdgeNodes(ctx context.Context) ([]EdgeNode, error) {
	url := fmt.Sprintf("%s/api/v1/edge/nodes", c.baseURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var nodes []EdgeNode
	if err := json.NewDecoder(resp.Body).Decode(&nodes); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return nodes, nil
}

// GetEdgeNode retrieves a specific edge node by ID
func (c *Client) GetEdgeNode(ctx context.Context, nodeID string) (*EdgeNode, error) {
	if nodeID == "" {
		return nil, errors.New("nodeID is required")
	}

	url := fmt.Sprintf("%s/api/v1/edge/nodes/%s", c.baseURL, nodeID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("edge node not found: %s", nodeID)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var node EdgeNode
	if err := json.NewDecoder(resp.Body).Decode(&node); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &node, nil
}

// ReportEdgeNodeHealth sends edge node health status
func (c *Client) ReportEdgeNodeHealth(ctx context.Context, health *HealthCheck) error {
	if health == nil {
		return errors.New("health check cannot be nil")
	}

	url := fmt.Sprintf("%s/api/v1/edge/health", c.baseURL)

	payload, err := json.Marshal(health)
	if err != nil {
		return fmt.Errorf("failed to marshal health check: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Body = io.NopCloser(io.Reader(bytes.NewReader(payload)))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// ConnectToEdge establishes connection to an edge node
func (c *Client) ConnectToEdge(ctx context.Context, nodeID string, sessionID string) error {
	if nodeID == "" || sessionID == "" {
		return errors.New("nodeID and sessionID are required")
	}

	url := fmt.Sprintf("%s/api/v1/edge/connect", c.baseURL)

	payload := map[string]string{
		"nodeId":    nodeID,
		"sessionId": sessionID,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Body = io.NopCloser(io.Reader(bytes.NewReader(data)))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// SetTimeout sets the client timeout
func (c *Client) SetTimeout(timeout time.Duration) {
	c.timeout = timeout
	c.httpClient.Timeout = timeout
}

// Health checks the API endpoint availability
func (c *Client) Health(ctx context.Context) error {
	url := fmt.Sprintf("%s/health", c.baseURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("API unreachable: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	return nil
}
