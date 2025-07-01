package rustpbx

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/google/uuid"
)

// Client represents the RustPBX WebSocket client
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a new RustPBX client
func NewClient(baseURL string) *Client {
	// Ensure baseURL doesn't end with a slash
	baseURL = strings.TrimSuffix(baseURL, "/")
	
	return &Client{
		baseURL:    baseURL,
		httpClient: &http.Client{},
	}
}

// NewClientWithHTTPClient creates a new RustPBX client with a custom HTTP client
func NewClientWithHTTPClient(baseURL string, httpClient *http.Client) *Client {
	baseURL = strings.TrimSuffix(baseURL, "/")
	
	return &Client{
		baseURL:    baseURL,
		httpClient: httpClient,
	}
}

// ConnectCall establishes a WebSocket connection to the /call endpoint
func (c *Client) ConnectCall(ctx context.Context, options *ConnectionOptions) (*Connection, error) {
	return c.connectWebSocket(ctx, "/call", options)
}

// ConnectWebRTC establishes a WebSocket connection to the /call/webrtc endpoint
func (c *Client) ConnectWebRTC(ctx context.Context, options *ConnectionOptions) (*Connection, error) {
	return c.connectWebSocket(ctx, "/call/webrtc", options)
}

// ConnectSIP establishes a WebSocket connection to the /call/sip endpoint
func (c *Client) ConnectSIP(ctx context.Context, options *ConnectionOptions) (*Connection, error) {
	return c.connectWebSocket(ctx, "/call/sip", options)
}

// connectWebSocket is the internal method to establish WebSocket connections
func (c *Client) connectWebSocket(ctx context.Context, endpoint string, options *ConnectionOptions) (*Connection, error) {
	if options == nil {
		options = &ConnectionOptions{}
	}

	// Generate session ID if not provided
	sessionID := options.SessionID
	if sessionID == "" {
		sessionID = uuid.New().String()
	}

	// Build WebSocket URL
	wsURL, err := c.buildWebSocketURL(endpoint, sessionID, options.Dump)
	if err != nil {
		return nil, fmt.Errorf("failed to build WebSocket URL: %w", err)
	}

	// Create and return connection
	conn, err := NewConnection(ctx, wsURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create WebSocket connection: %w", err)
	}

	return conn, nil
}

// buildWebSocketURL builds the WebSocket URL with query parameters
func (c *Client) buildWebSocketURL(endpoint, sessionID string, dump bool) (string, error) {
	// Convert HTTP(S) URL to WebSocket URL
	wsURL := c.baseURL
	if strings.HasPrefix(wsURL, "http://") {
		wsURL = strings.Replace(wsURL, "http://", "ws://", 1)
	} else if strings.HasPrefix(wsURL, "https://") {
		wsURL = strings.Replace(wsURL, "https://", "wss://", 1)
	} else if !strings.HasPrefix(wsURL, "ws://") && !strings.HasPrefix(wsURL, "wss://") {
		// Assume ws:// if no protocol specified
		wsURL = "ws://" + wsURL
	}

	// Parse URL to add query parameters
	u, err := url.Parse(wsURL + endpoint)
	if err != nil {
		return "", err
	}

	// Add query parameters
	q := u.Query()
	if sessionID != "" {
		q.Set("id", sessionID)
	}
	q.Set("dump", fmt.Sprintf("%t", dump))
	u.RawQuery = q.Encode()

	return u.String(), nil
}

// GetActiveCalls retrieves a list of all currently active calls
func (c *Client) GetActiveCalls(ctx context.Context) (*CallListResponse, error) {
	url := c.baseURL + "/call/lists"
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	
	var result CallListResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	return &result, nil
}

// KillCall forcefully terminates an active call by ID
func (c *Client) KillCall(ctx context.Context, callID string) error {
	url := c.baseURL + "/call/kill/" + callID
	
	req, err := http.NewRequestWithContext(ctx, "POST", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("call with ID %s not found", callID)
	}
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	
	return nil
}

// GetICEServers retrieves ICE servers configuration for WebRTC connections
func (c *Client) GetICEServers(ctx context.Context) ([]ICEServer, error) {
	url := c.baseURL + "/iceservers"
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	
	var result []ICEServer
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	return result, nil
}

// ProxyLLMRequest forwards a request to the LLM proxy endpoint
func (c *Client) ProxyLLMRequest(ctx context.Context, path string, method string, body io.Reader, headers map[string]string) (*http.Response, error) {
	url := c.baseURL + "/llm/v1/" + strings.TrimPrefix(path, "/")
	
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	// Set default headers
	req.Header.Set("Content-Type", "application/json")
	
	// Set custom headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	
	return resp, nil
}