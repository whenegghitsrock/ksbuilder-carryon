package chartmuseum

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// UploadResult holds the result of an upload.
type UploadResult struct {
	Saved bool   `json:"saved"`
	ID    string `json:"id,omitempty"`
}

// Client is the ChartMuseum HTTP client.
type Client struct {
	BaseURL  string
	Username string
	Password string
	Client   *http.Client
}

// NewClient creates a new client.
func NewClient(baseURL, username, password string) *Client {
	baseURL = strings.TrimSuffix(baseURL, "/")
	return &Client{
		BaseURL:  baseURL,
		Username: username,
		Password: password,
		Client:   &http.Client{},
	}
}

// UploadChart uploads a .tgz to ChartMuseum via POST /api/charts (multipart form chart=@file).
func (c *Client) UploadChart(tgzPath string) (*UploadResult, error) {
	f, err := os.Open(tgzPath)
	if err != nil {
		return nil, fmt.Errorf("open chart file: %w", err)
	}
	defer f.Close()

	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)

	part, err := w.CreateFormFile("chart", filepath.Base(tgzPath))
	if err != nil {
		return nil, err
	}
	if _, err := io.Copy(part, f); err != nil {
		return nil, err
	}
	if err := w.Close(); err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, c.BaseURL+"/api/charts", &buf)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(c.Username, c.Password)
	req.Header.Set("Content-Type", w.FormDataContentType())

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("chartmuseum request failed: %w\nHint: check repo URL and network", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("chartmuseum returned %d: %s\nHint: check --username and --password for 401/403, or --repo URL for 404", resp.StatusCode, string(body))
	}

	// Skip JSON parsing; only check status code.
	return &UploadResult{Saved: true}, nil
}
