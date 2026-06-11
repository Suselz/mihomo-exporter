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
	httpClient      *http.Client
	baseURL         string
	connectionsPath string
}

func NewClient(baseURL, path string) *Client {
	baseURL = strings.TrimRight(baseURL, "/")
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return &Client{
		httpClient:      &http.Client{Timeout: 10 * time.Second},
		baseURL:         baseURL,
		connectionsPath: path,
	}
}

func (c *Client) Fetch(ctx context.Context) (Snapshot, error) {
	var snap Snapshot
	if err := c.fetchJSON(ctx, c.connectionsPath, &snap); err != nil {
		return Snapshot{}, err
	}
	return snap, nil
}

func (c *Client) FetchMemory(ctx context.Context) (MemorySnapshot, error) {
	var snap MemorySnapshot
	if err := c.fetchJSON(ctx, "/memory", &snap); err != nil {
		return MemorySnapshot{}, err
	}
	return snap, nil
}

func (c *Client) FetchTraffic(ctx context.Context) (TrafficSnapshot, error) {
	var snap TrafficSnapshot
	if err := c.fetchJSON(ctx, "/traffic", &snap); err != nil {
		return TrafficSnapshot{}, err
	}
	return snap, nil
}

func (c *Client) FetchVersion(ctx context.Context) (VersionSnapshot, error) {
	var snap VersionSnapshot
	if err := c.fetchJSON(ctx, "/version", &snap); err != nil {
		return VersionSnapshot{}, err
	}
	return snap, nil
}

func (c *Client) FetchProxies(ctx context.Context) (ProxiesSnapshot, error) {
	var snap ProxiesSnapshot
	if err := c.fetchJSON(ctx, "/proxies", &snap); err != nil {
		return ProxiesSnapshot{}, err
	}
	return snap, nil
}

func (c *Client) FetchRules(ctx context.Context) (RulesSnapshot, error) {
	var snap RulesSnapshot
	if err := c.fetchJSON(ctx, "/rules", &snap); err != nil {
		return RulesSnapshot{}, err
	}
	return snap, nil
}

func (c *Client) fetchJSON(ctx context.Context, path string, out any) error {
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	url := c.baseURL + path

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status for %s: %s", path, resp.Status)
	}

	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("%w: %v", ErrDecode, err)
	}
	return nil
}
