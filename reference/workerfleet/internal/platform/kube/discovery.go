package kube

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	"github.com/spcent/plumego/reference/workerfleet/internal/domain"
)

const (
	defaultServiceAccountToken = "/var/run/secrets/kubernetes.io/serviceaccount/token"
	defaultNamespacePath       = "/var/run/secrets/kubernetes.io/serviceaccount/namespace"
)

type Config struct {
	APIHost         string
	BearerToken     string
	Namespace       string
	LabelSelector   string
	WorkerContainer string
	HTTPClient      *http.Client
}

type Client struct {
	cfg        Config
	httpClient *http.Client
}

type listMeta struct {
	ResourceVersion string `json:"resourceVersion"`
}

type PodList struct {
	Items    []Pod    `json:"items"`
	Metadata listMeta `json:"metadata"`
}

type InventorySync struct {
	client          *Client
	snapshots       domain.SnapshotStore
	now             func() time.Time
	workerContainer string
	policy          domain.StatusPolicy
}

func NewClient(cfg Config) (*Client, error) {
	if strings.TrimSpace(cfg.APIHost) == "" {
		cfg.APIHost = defaultAPIHost()
	}
	if strings.TrimSpace(cfg.Namespace) == "" {
		cfg.Namespace = defaultNamespace()
	}
	if strings.TrimSpace(cfg.BearerToken) == "" {
		token, err := os.ReadFile(defaultServiceAccountToken)
		if err == nil {
			cfg.BearerToken = strings.TrimSpace(string(token))
		}
	}

	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: 15 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{MinVersion: tls.VersionTLS12},
			},
		}
	}

	return &Client{
		cfg:        cfg,
		httpClient: httpClient,
	}, nil
}

func (c *Client) ListPods(ctx context.Context) (PodList, error) {
	endpoint, err := c.podEndpoint(false, "")
	if err != nil {
		return PodList{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return PodList{}, err
	}
	c.authorize(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return PodList{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return PodList{}, fmt.Errorf("list pods: status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var out PodList
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return PodList{}, err
	}
	return out, nil
}

func NewInventorySync(client *Client, snapshots domain.SnapshotStore, workerContainer string, policy domain.StatusPolicy) *InventorySync {
	if policy == (domain.StatusPolicy{}) {
		policy = domain.DefaultStatusPolicy()
	}
	return &InventorySync{
		client:          client,
		snapshots:       snapshots,
		now:             time.Now().UTC,
		workerContainer: workerContainer,
		policy:          policy,
	}
}

func (s *InventorySync) SyncOnce(ctx context.Context) (string, error) {
	if s.client == nil {
		return "", fmt.Errorf("kubernetes client is required")
	}
	if s.snapshots == nil {
		return "", fmt.Errorf("snapshot store is required")
	}

	list, err := s.client.ListPods(ctx)
	if err != nil {
		return "", err
	}
	for _, pod := range list.Items {
		identity, podSnapshot, ok := MapPod(pod, s.workerContainer)
		if !ok {
			continue
		}
		previous, found, err := s.snapshots.GetWorkerSnapshot(identity.WorkerID)
		if err != nil {
			return "", err
		}
		if !found {
			previous = domain.WorkerSnapshot{}
		}
		merged, _ := domain.ReconcilePodSnapshot(previous, identity, podSnapshot, s.now(), s.policy)
		if err := s.snapshots.UpsertWorkerSnapshot(merged); err != nil {
			return "", err
		}
	}
	return list.Metadata.ResourceVersion, nil
}

func defaultAPIHost() string {
	host := strings.TrimSpace(os.Getenv("KUBERNETES_SERVICE_HOST"))
	port := strings.TrimSpace(os.Getenv("KUBERNETES_SERVICE_PORT_HTTPS"))
	if host == "" {
		return ""
	}
	if port == "" {
		port = "443"
	}
	return "https://" + host + ":" + port
}

func defaultNamespace() string {
	data, err := os.ReadFile(defaultNamespacePath)
	if err != nil {
		return "default"
	}
	namespace := strings.TrimSpace(string(data))
	if namespace == "" {
		return "default"
	}
	return namespace
}

func (c *Client) podEndpoint(watch bool, resourceVersion string) (string, error) {
	if strings.TrimSpace(c.cfg.APIHost) == "" {
		return "", fmt.Errorf("kubernetes API host is required")
	}
	base, err := url.Parse(c.cfg.APIHost)
	if err != nil {
		return "", err
	}

	base.Path = path.Join(base.Path, "/api/v1/namespaces", c.cfg.Namespace, "pods")
	query := base.Query()
	if c.cfg.LabelSelector != "" {
		query.Set("labelSelector", c.cfg.LabelSelector)
	}
	if watch {
		query.Set("watch", "1")
	}
	if resourceVersion != "" {
		query.Set("resourceVersion", resourceVersion)
	}
	base.RawQuery = query.Encode()
	return base.String(), nil
}

func (c *Client) authorize(req *http.Request) {
	if strings.TrimSpace(c.cfg.BearerToken) == "" {
		return
	}
	req.Header.Set("Authorization", "Bearer "+c.cfg.BearerToken)
}
