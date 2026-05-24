package kube

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

var ErrResourceVersionExpired = errors.New("kubernetes resource version expired")

type WatchEvent struct {
	Type   string           `json:"type"`
	Object Pod              `json:"object,omitempty"`
	Status KubernetesStatus `json:"status,omitempty"`
}

type KubernetesStatus struct {
	Code    int    `json:"code,omitempty"`
	Reason  string `json:"reason,omitempty"`
	Message string `json:"message,omitempty"`
}

func (e *WatchEvent) UnmarshalJSON(data []byte) error {
	var raw struct {
		Type   string          `json:"type"`
		Object json.RawMessage `json:"object"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	e.Type = raw.Type
	if strings.EqualFold(raw.Type, "ERROR") {
		return json.Unmarshal(raw.Object, &e.Status)
	}
	if len(raw.Object) == 0 {
		return nil
	}
	return json.Unmarshal(raw.Object, &e.Object)
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
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("watch pods: status %d", resp.StatusCode)
	}

	decoder := json.NewDecoder(resp.Body)
	for {
		var event WatchEvent
		if err := decoder.Decode(&event); err != nil {
			if err == io.EOF {
				return nil
			}
			if ctx.Err() != nil {
				return nil
			}
			return err
		}
		if strings.EqualFold(event.Type, "ERROR") {
			if event.Status.Code == http.StatusGone || strings.EqualFold(event.Status.Reason, "Expired") {
				return ErrResourceVersionExpired
			}
			return fmt.Errorf("watch pods: kubernetes error %d %s", event.Status.Code, kubernetesStatusReason(event.Status))
		}
		if err := onEvent(event); err != nil {
			return err
		}
	}
}

func kubernetesStatusReason(status KubernetesStatus) string {
	reason := strings.TrimSpace(status.Reason)
	if reason == "" {
		return "unknown"
	}
	return reason
}
