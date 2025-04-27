package gocache

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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

// Get retrieves a string value by key
func (c *Client) Get(key string) (string, error) {
	url := fmt.Sprintf("%s/string/%s", c.BaseURL, key)
	resp, err := c.client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", c.parseError(resp.Body)
	}

	response, err := c.parseResponse(resp.Body)
	if err != nil {
		return "", err
	}

	var result string
	err = json.Unmarshal(response.Data, &result)
	if err != nil {
		return "", fmt.Errorf("error parsing response data: %w", err)
	}

	return result, nil
}

// Set sets a string value with optional TTL
func (c *Client) Set(key, value string, ttl time.Duration) error {
	url := fmt.Sprintf("%s/string/%s", c.BaseURL, key)
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
	url := fmt.Sprintf("%s/string/%s", c.BaseURL, key)

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
	url := fmt.Sprintf("%s/string/%s", c.BaseURL, key)

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
	url := fmt.Sprintf("%s/list/%s", c.BaseURL, key)
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
	url := fmt.Sprintf("%s/list/%s", c.BaseURL, key)

	resp, err := c.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.parseError(resp.Body)
	}

	response, err := c.parseResponse(resp.Body)
	if err != nil {
		return nil, err
	}

	var result []string
	err = json.Unmarshal(response.Data, &result)
	if err != nil {
		return nil, fmt.Errorf("error parsing response data: %w", err)
	}

	return result, nil
}

// Push adds a value to the end of a list
func (c *Client) Push(key, value string) error {
	url := fmt.Sprintf("%s/list/%s/push", c.BaseURL, key)

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

// Pop removes and returns the last value from a list
func (c *Client) Pop(key string) (string, error) {
	url := fmt.Sprintf("%s/list/%s/pop", c.BaseURL, key)

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return "", err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", c.parseError(resp.Body)
	}

	response, err := c.parseResponse(resp.Body)
	if err != nil {
		return "", err
	}

	var result string
	err = json.Unmarshal(response.Data, &result)
	if err != nil {
		return "", fmt.Errorf("error parsing response data: %w", err)
	}

	return result, nil
}

// RemoveList deletes a list
func (c *Client) RemoveList(key string) error {
	url := fmt.Sprintf("%s/list/%s", c.BaseURL, key)

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
	url := fmt.Sprintf("%s/ttl/%s", c.BaseURL, key)

	resp, err := c.client.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, c.parseError(resp.Body)
	}

	response, err := c.parseResponse(resp.Body)
	if err != nil {
		return 0, err
	}

	var ttlSeconds int64
	err = json.Unmarshal(response.Data, &ttlSeconds)
	if err != nil {
		return 0, fmt.Errorf("error parsing response data: %w", err)
	}

	if ttlSeconds < 0 {
		return -1, nil // No expiration
	}

	return time.Duration(ttlSeconds) * time.Second, nil
}

// SetTTL sets or updates the TTL for a key
func (c *Client) SetTTL(key string, ttl time.Duration) error {
	ttlSeconds := int(ttl.Seconds())
	url := fmt.Sprintf("%s/ttl/%s?ttl=%d", c.BaseURL, key, ttlSeconds)

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
