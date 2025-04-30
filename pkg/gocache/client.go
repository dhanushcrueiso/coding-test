package gocache

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// Client represents a client for the GoCache server
type Client struct {
	BaseURL string
	client  *http.Client
}

// NewClient creates a new GoCache client
func NewClient(baseURL string) *Client {
	return &Client{
		BaseURL: baseURL,
		client: &http.Client{
			Timeout: time.Second * 10,
		},
	}
}

// responseBody is the structure for parsing server responses
type responseBody struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data,omitempty"`
	Error   string          `json:"error,omitempty"`
}

type dataBody struct {
	Key   string      `json:"key"`
	Type  int         `json:"type"`
	Value interface{} `json:"value"`
	// Add any other fields that Fiber might include
}

type dataresponselist struct {
	Data    string `json:"value"`
	Message string `json:"messsage"`
}

// Get retrieves a string value by key
func (c *Client) Get(key string) (string, error) {
	url := fmt.Sprintf("%s/api/strings/%s", c.BaseURL, key)
	resp, err := c.client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("server returned status %d: %s",
			resp.StatusCode, string(bodyBytes))
	}

	// Parse the response
	var response dataBody
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("error parsing response: %w (body: %s)",
			err, string(bodyBytes))
	}

	// Convert the value to string
	value, ok := response.Value.(string)
	if !ok {
		return "", fmt.Errorf("expected string value, got: %T", response.Value)
	}

	return value, nil
}

// Set sets a string value with optional TTL
func (c *Client) Set(key, value string, ttl time.Duration) error {
	url := fmt.Sprintf("%s/api/strings/%s", c.BaseURL, key)
	if ttl > 0 {
		url = fmt.Sprintf("%s?ttl=%d", url, int(ttl.Seconds()))
	}

	data := struct {
		Value string `json:"value"`
	}{
		Value: value,
	}

	body, err := json.Marshal(data)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return c.parseError(resp.Body)
	}

	return nil
}

// Update updates an existing string value
func (c *Client) Update(key, value string) error {
	url := fmt.Sprintf("%s/api/strings/%s", c.BaseURL, key)

	data := struct {
		Value string `json:"value"`
	}{
		Value: value,
	}

	body, err := json.Marshal(data)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return c.parseError(resp.Body)
	}

	return nil
}

// Remove deletes a key
func (c *Client) Remove(key string) error {
	url := fmt.Sprintf("%s/api/strings/%s", c.BaseURL, key)

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return c.parseError(resp.Body)
	}

	return nil
}

// CreateList initializes a new list with optional TTL
func (c *Client) CreateList(key string, ttl time.Duration) error {
	url := fmt.Sprintf("%s/api/list/%s", c.BaseURL, key)
	if ttl > 0 {
		url = fmt.Sprintf("%s?ttl=%d", url, int(ttl.Seconds()))
	}

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return c.parseError(resp.Body)
	}

	return nil
}

// GetList retrieves all items in a list
func (c *Client) GetList(key string) ([]string, error) {
	url := fmt.Sprintf("%s/api/list/%s", c.BaseURL, key)

	resp, err := c.client.Get(url)
	if err != nil {
		fmt.Println("check oje")
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println("check two")
		return nil, c.parseError(resp.Body)
	}

	// Parse the response
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	// Parse the Fiber response directly first
	var fiberResponse struct {
		Data    json.RawMessage `json:"data"`
		Message string          `json:"message"`
	}

	if err := json.Unmarshal(bodyBytes, &fiberResponse); err != nil {
		// If direct parsing fails, fall back to parseResponse
		resp.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		response, err := c.parseResponse(resp.Body)
		if err != nil {
			return nil, err
		}

		var result []string
		if err := json.Unmarshal(response.Data, &result); err != nil {
			return nil, fmt.Errorf("error parsing response data: %w", err)
		}

		return result, nil
	}

	// If direct parsing succeeded, handle the data field
	var result []string
	if err := json.Unmarshal(fiberResponse.Data, &result); err != nil {
		return nil, fmt.Errorf("error parsing data array: %w", err)
	}

	return result, nil
}

// Push adds a value to the end of a list
func (c *Client) Push(key, value string) error {
	url := fmt.Sprintf("%s/api/list/%s/push", c.BaseURL, key)

	data := struct {
		Value string `json:"value"`
	}{
		Value: value,
	}

	body, err := json.Marshal(data)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PATCH", url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return c.parseError(resp.Body)
	}

	return nil
}

type PopResponse struct {
	Data    string `json:"data"`    // The popped string value
	Message string `json:"message"` // Success message
}

// Pop removes and returns the last value from a list
func (c *Client) Pop(key string) (string, error) {
	url := fmt.Sprintf("%s/api/list/%s/pop", c.BaseURL, key)

	req, err := http.NewRequest("PATCH", url, nil)
	if err != nil {
		return "", fmt.Errorf("error creating request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read the response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %w", err)
	}

	// Log the raw response for debugging
	fmt.Printf("Pop response: %s\n", string(bodyBytes))

	if resp.StatusCode != http.StatusOK {
		// Create a new reader for the parseError function
		resp.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		return "", c.parseError(resp.Body)
	}

	// Parse the response into our structure
	var popResponse PopResponse
	if err := json.Unmarshal(bodyBytes, &popResponse); err != nil {
		return "", fmt.Errorf("error parsing response: %w (body: %s)",
			err, string(bodyBytes))
	}

	// Check if we got a valid string value
	if popResponse.Data == "" {
		// Some APIs might return null/empty on an empty list
		return "", fmt.Errorf("list is empty or returned empty value")
	}

	return popResponse.Data, nil
}

// RemoveList deletes a list
func (c *Client) RemoveList(key string) error {
	url := fmt.Sprintf("%s/api/list/%s", c.BaseURL, key)

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return c.parseError(resp.Body)
	}

	return nil
}

// GetTTL returns the remaining TTL for a key
func (c *Client) GetTTL(key string) (time.Duration, error) {
	url := fmt.Sprintf("%s/api/ttl/%s", c.BaseURL, key)

	resp, err := c.client.Get(url)
	if err != nil {
		return 0, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read the response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("error reading response: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("server returned error (status %d): %s",
			resp.StatusCode, string(bodyBytes))
	}

	// Parse the response
	var response struct {
		Message string  `json:"message"`
		Data    float64 `json:"data"`
	}

	if err := json.Unmarshal(bodyBytes, &response); err != nil {
		return 0, fmt.Errorf("error parsing response: %w", err)
	}

	// Check for negative value indicating no expiration
	if response.Data < 0 {
		return -1, nil // No expiration
	}

	// Convert seconds to duration
	return time.Duration(response.Data * float64(time.Second)), nil
}

// SetTTL sets or updates the TTL for a key
func (c *Client) SetTTL(key string, ttl time.Duration) error {
	ttlSeconds := int(ttl.Seconds())

	// Build URL with properly encoded query parameter
	baseURL := fmt.Sprintf("%s/api/ttl/%s", c.BaseURL, key)
	reqURL, err := url.Parse(baseURL)
	if err != nil {
		return fmt.Errorf("error parsing URL: %w", err)
	}

	// Add query parameters
	query := reqURL.Query()
	query.Set("ttl", strconv.Itoa(ttlSeconds))
	reqURL.RawQuery = query.Encode()

	// Create request
	req, err := http.NewRequest(http.MethodPost, reqURL.String(), nil)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	// Set content type
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read the response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response: %w", err)
	}

	// Check if successful
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned error (status %d): %s",
			resp.StatusCode, string(bodyBytes))
	}

	return nil
}

// Helper methods

func (c *Client) parseResponse(body io.Reader) (*responseBody, error) {
	var response responseBody
	err := json.NewDecoder(body).Decode(&response)
	if err != nil {
		return nil, fmt.Errorf("error parsing response: %w", err)
	}

	if !response.Success {
		return nil, fmt.Errorf("server returned error: %s", response.Error)
	}

	return &response, nil
}

func (c *Client) parseError(body io.Reader) error {
	var response responseBody
	err := json.NewDecoder(body).Decode(&response)
	if err != nil {
		return fmt.Errorf("error parsing error response: %w", err)
	}

	if response.Error != "" {
		return fmt.Errorf("server error: %s", response.Error)
	}

	return fmt.Errorf("unknown server error")
}
