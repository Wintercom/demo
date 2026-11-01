package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client 封装 http.Client
type Client struct {
	baseURL string
	http    *http.Client
}

// NewClient 创建客户端
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		http: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// GetHello 调用 /hello
func (c *Client) GetHello() (string, error) {
	resp, err := c.http.Get(c.baseURL + "/hello")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

// GetUser 调用 /users/{id}
func (c *Client) GetUser(id string) (map[string]interface{}, error) {
	resp, err := c.http.Get(fmt.Sprintf("%s/users/%s", c.baseURL, id))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var user map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
	}
	return user, nil
}

// PostEcho 调用 /echo
func (c *Client) PostEcho(payload map[string]string) (map[string]string, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	resp, err := c.http.Post(c.baseURL+"/echo", "application/json", bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var result map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result, nil
}