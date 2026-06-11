package mihomo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

var ErrDecode = errors.New("decode error")

type Client struct {
	httpClient *http.Client
	url        string
}

func NewClient(baseURL, path string) *Client {
	baseURL = strings.TrimRight(baseURL, "/")
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return &Client{
		httpClient: &http.Client{Timeout: 10 * time.Second},
		url:        baseURL + path,
	}
}

func (c *Client) Fetch(ctx context.Context) (Snapshot, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url, nil)
	if err != nil {
		return Snapshot{}, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return Snapshot{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Snapshot{}, fmt.Errorf("unexpected status: %s", resp.Status)
	}

	var snap Snapshot
	if err := json.NewDecoder(resp.Body).Decode(&snap); err != nil {
		return Snapshot{}, fmt.Errorf("%w: %v", ErrDecode, err)
	}
	return snap, nil
}
