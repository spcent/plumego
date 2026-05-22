package kube

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	"workerfleet/internal/domain"
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
	metrics         InventoryMetricsObserver
}

type InventoryMetricsObserver interface {
	ObserveInventorySync(snapshots []domain.WorkerSnapshot, operation string, result string, duration time.Duration)
}

type InventorySyncOption func(*InventorySync)

func WithMetricsObserver(observer InventoryMetricsObserver) InventorySyncOption {
	return func(s *InventorySync) {
		s.metrics = observer
	}
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
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 4096))
		return PodList{}, fmt.Errorf("list pods: status %d", resp.StatusCode)
	}

	var out PodList
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return PodList{}, err
	}
	return out, nil
}

func NewInventorySync(client *Client, snapshots domain.SnapshotStore, workerContainer string, policy domain.StatusPolicy, opts ...InventorySyncOption) *InventorySync {
	if policy == (domain.StatusPolicy{}) {
		policy = domain.DefaultStatusPolicy()
	}
	syncer := &InventorySync{
		client:          client,
		snapshots:       snapshots,
		now:             time.Now().UTC,
		workerContainer: workerContainer,
		policy:          policy,
	}
	for _, opt := range opts {
		if opt != nil {
			opt(syncer)
		}
	}
	return syncer
}

func (s *InventorySync) SyncOnce(ctx context.Context) (resourceVersion string, err error) {
	started := time.Now()
	observed := make([]domain.WorkerSnapshot, 0)
	defer func() {
		if s.metrics == nil {
			return
		}
		result := "success"
		if err != nil {
			result = "error"
		}
		s.metrics.ObserveInventorySync(observed, "sync_once", result, time.Since(started))
	}()

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
		merged, ok, err := s.applyPod(ctx, pod)
		if err != nil {
			return "", err
		}
		if ok {
			observed = append(observed, merged)
		}
	}
	return list.Metadata.ResourceVersion, nil
}

func (s *InventorySync) SyncWatch(ctx context.Context) (string, error) {
	resourceVersion, err := s.SyncOnce(ctx)
	if err != nil {
		return "", err
	}
	nextResourceVersion, err := s.Watch(ctx, resourceVersion)
	if errors.Is(err, ErrResourceVersionExpired) {
		return s.SyncOnce(ctx)
	}
	if err != nil {
		return "", err
	}
	return nextResourceVersion, nil
}

func (s *InventorySync) Watch(ctx context.Context, resourceVersion string) (string, error) {
	if s.client == nil {
		return "", fmt.Errorf("kubernetes client is required")
	}
	if s.snapshots == nil {
		return "", fmt.Errorf("snapshot store is required")
	}
	latest := resourceVersion
	err := s.client.WatchPods(ctx, resourceVersion, func(event WatchEvent) error {
		switch strings.ToUpper(strings.TrimSpace(event.Type)) {
		case "ADDED", "MODIFIED":
			merged, ok, err := s.applyPod(ctx, event.Object)
			if err != nil || !ok {
				return err
			}
			if event.Object.Metadata.ResourceVersion != "" {
				latest = event.Object.Metadata.ResourceVersion
			}
			_ = merged
			return nil
		case "DELETED":
			if err := s.applyDeletedPod(ctx, event.Object); err != nil {
				return err
			}
			if event.Object.Metadata.ResourceVersion != "" {
				latest = event.Object.Metadata.ResourceVersion
			}
			return nil
		case "BOOKMARK":
			if event.Object.Metadata.ResourceVersion != "" {
				latest = event.Object.Metadata.ResourceVersion
			}
			return nil
		case "":
			return nil
		default:
			return fmt.Errorf("unsupported kubernetes watch event type %q", event.Type)
		}
	})
	if err != nil {
		return "", err
	}
	return latest, nil
}

func (s *InventorySync) applyPod(ctx context.Context, pod Pod) (domain.WorkerSnapshot, bool, error) {
	identity, podSnapshot, ok := MapPod(pod, s.workerContainer)
	if !ok {
		return domain.WorkerSnapshot{}, false, nil
	}
	previous, found, err := s.snapshots.GetWorkerSnapshot(ctx, identity.WorkerID)
	if err != nil {
		return domain.WorkerSnapshot{}, false, err
	}
	if !found {
		previous = domain.WorkerSnapshot{}
	}
	merged, _ := domain.ReconcilePodSnapshot(previous, identity, podSnapshot, s.now(), s.policy)
	if err := s.snapshots.UpsertWorkerSnapshot(ctx, merged); err != nil {
		return domain.WorkerSnapshot{}, false, err
	}
	return merged, true, nil
}

func (s *InventorySync) applyDeletedPod(ctx context.Context, pod Pod) error {
	workerID := domain.WorkerID(strings.TrimSpace(pod.Metadata.Name))
	if workerID == "" {
		return nil
	}
	previous, found, err := s.snapshots.GetWorkerSnapshot(ctx, workerID)
	if err != nil {
		return err
	}
	identity := previous.Identity
	if !found {
		identity.WorkerID = workerID
		identity.PodName = strings.TrimSpace(pod.Metadata.Name)
		identity.Namespace = strings.TrimSpace(pod.Metadata.Namespace)
		identity.PodUID = domain.PodUID(strings.TrimSpace(pod.Metadata.UID))
		identity.NodeName = strings.TrimSpace(pod.Spec.NodeName)
	}
	deletedAt := parseTime(pod.Metadata.DeletionTimestamp)
	if deletedAt.IsZero() {
		deletedAt = s.now()
	}
	podSnapshot := previous.Pod
	podSnapshot.DeletedAt = deletedAt
	merged, _ := domain.ReconcilePodSnapshot(previous, identity, podSnapshot, s.now(), s.policy)
	return s.snapshots.UpsertWorkerSnapshot(ctx, merged)
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
