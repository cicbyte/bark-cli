package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
	auth       *AuthConfig
}

type AuthConfig struct {
	Username string
	Password string
}

type CommonResp struct {
	Code      int    `json:"code"`
	Message   string `json:"message"`
	Data      any    `json:"data,omitempty"`
	Timestamp int64  `json:"timestamp"`
}

type ServerInfo struct {
	Version string `json:"version"`
	Build   string `json:"build"`
	Arch    string `json:"arch"`
	Commit  string `json:"commit"`
	Devices int    `json:"devices"`
}

type RegisterResp struct {
	Key         string `json:"key"`
	DeviceKey   string `json:"device_key"`
	DeviceToken string `json:"device_token"`
}

type PushRequest struct {
	DeviceKey  string   `json:"device_key,omitempty"`
	DeviceKeys []string `json:"device_keys,omitempty"`
	Title      string   `json:"title,omitempty"`
	Subtitle   string   `json:"subtitle,omitempty"`
	Body       string   `json:"body,omitempty"`
	Sound      string   `json:"sound,omitempty"`
	Level      string   `json:"level,omitempty"`
	Volume     int      `json:"volume,omitempty"`
	Badge      int      `json:"badge,omitempty"`
	Icon       string   `json:"icon,omitempty"`
	Group      string   `json:"group,omitempty"`
	URL        string   `json:"url,omitempty"`
	Copy       string   `json:"copy,omitempty"`
	IsArchive  string   `json:"isArchive,omitempty"`
	Call       string   `json:"call,omitempty"`
	ID         string   `json:"id,omitempty"`
	Image      string   `json:"image,omitempty"`
	Ciphertext string   `json:"ciphertext,omitempty"`
	Action     string   `json:"action,omitempty"`
}

type BatchResult struct {
	Code      int    `json:"code"`
	Message   string `json:"message,omitempty"`
	DeviceKey string `json:"device_key"`
}

func New(serverURL string, timeout int, auth *AuthConfig) *Client {
	if timeout <= 0 {
		timeout = 10
	}
	return &Client{
		baseURL: strings.TrimRight(serverURL, "/"),
		httpClient: &http.Client{
			Timeout: time.Duration(timeout) * time.Second,
		},
		auth: auth,
	}
}

func (c *Client) newRequest(method, path string, body any) (*http.Request, error) {
	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request: %w", err)
		}
		reqBody = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, c.baseURL+path, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.auth != nil && c.auth.Username != "" && c.auth.Password != "" {
		req.SetBasicAuth(c.auth.Username, c.auth.Password)
	}
	return req, nil
}

func (c *Client) doRequest(req *http.Request) (*CommonResp, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Handle HTTP 418 (auth failed, non-JSON response)
	if resp.StatusCode == 418 {
		return nil, fmt.Errorf("authentication failed (HTTP 418)")
	}

	var result CommonResp
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w (body: %s)", err, string(body))
	}
	return &result, nil
}

func (c *Client) doRequestRaw(req *http.Request) ([]byte, int, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("failed to read response: %w", err)
	}
	return body, resp.StatusCode, nil
}

// Push sends a notification to a single device.
func (c *Client) Push(req *PushRequest) (*CommonResp, error) {
	httpReq, err := c.newRequest("POST", "/push", req)
	if err != nil {
		return nil, err
	}
	return c.doRequest(httpReq)
}

// PushBatch sends notifications to multiple devices.
func (c *Client) PushBatch(req *PushRequest) (*CommonResp, error) {
	httpReq, err := c.newRequest("POST", "/push", req)
	if err != nil {
		return nil, err
	}
	return c.doRequest(httpReq)
}

// Register registers a device. Returns the device key.
func (c *Client) Register(deviceKey, deviceToken string) (*RegisterResp, error) {
	body := map[string]string{
		"device_token": deviceToken,
	}
	if deviceKey != "" {
		body["device_key"] = deviceKey
	}

	httpReq, err := c.newRequest("POST", "/register", body)
	if err != nil {
		return nil, err
	}
	resp, err := c.doRequest(httpReq)
	if err != nil {
		return nil, err
	}
	if resp.Code != 200 {
		return nil, fmt.Errorf("register failed (code %d): %s", resp.Code, resp.Message)
	}

	data, ok := resp.Data.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid response data")
	}

	result := &RegisterResp{
		Key:         strVal(data["key"]),
		DeviceKey:   strVal(data["device_key"]),
		DeviceToken: strVal(data["device_token"]),
	}
	return result, nil
}

// CheckDevice checks if a device is registered.
func (c *Client) CheckDevice(deviceKey string) error {
	httpReq, err := c.newRequest("GET", "/register/"+url.PathEscape(deviceKey), nil)
	if err != nil {
		return err
	}
	resp, err := c.doRequest(httpReq)
	if err != nil {
		return err
	}
	if resp.Code != 200 {
		return fmt.Errorf("device not registered (code %d): %s", resp.Code, resp.Message)
	}
	return nil
}

// Ping checks server connectivity.
func (c *Client) Ping() error {
	httpReq, err := c.newRequest("GET", "/ping", nil)
	if err != nil {
		return err
	}
	resp, err := c.doRequest(httpReq)
	if err != nil {
		return err
	}
	if resp.Code != 200 {
		return fmt.Errorf("ping failed (code %d): %s", resp.Code, resp.Message)
	}
	return nil
}

// Info retrieves server information.
func (c *Client) Info() (*ServerInfo, error) {
	httpReq, err := c.newRequest("GET", "/info", nil)
	if err != nil {
		return nil, err
	}
	body, statusCode, err := c.doRequestRaw(httpReq)
	if err != nil {
		return nil, err
	}
	if statusCode != 200 {
		return nil, fmt.Errorf("info request failed (HTTP %d): %s", statusCode, string(body))
	}

	var info ServerInfo
	if err := json.Unmarshal(body, &info); err != nil {
		return nil, fmt.Errorf("failed to parse info response: %w", err)
	}
	return &info, nil
}

func strVal(v any) string {
	if v == nil {
		return ""
	}
	s, ok := v.(string)
	if ok {
		return s
	}
	return fmt.Sprintf("%v", v)
}
