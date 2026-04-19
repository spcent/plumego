package kube

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type WatchEvent struct {
	Type   string `json:"type"`
	Object Pod    `json:"object"`
}

func (c *Client) WatchPods(ctx context.Context, resourceVersion string, onEvent func(WatchEvent) error) error {
	if onEvent == nil {
		return fmt.Errorf("watch callback is required")
	}

	endpoint, err := c.podEndpoint(true, resourceVersion)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}
	c.authorize(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("watch pods: status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	decoder := json.NewDecoder(resp.Body)
	for {
		var event WatchEvent
		if err := decoder.Decode(&event); err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		if err := onEvent(event); err != nil {
			return err
		}
	}
}
