package kube

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/spcent/plumego/reference/workerfleet/internal/domain"
)

func TestMapPod(t *testing.T) {
	pod := Pod{
		Metadata: PodMetadata{
			Name:      "worker-1",
			Namespace: "sim",
			UID:       "uid-1",
		},
		Spec: PodSpec{
			NodeName: "node-a",
			Containers: []Container{{
				Name:  "worker",
				Image: "worker:v1",
			}},
		},
		Status: PodStatus{
			Phase:     "Running",
			PodIP:     "10.0.0.10",
			HostIP:    "192.168.1.20",
			StartTime: "2026-04-19T14:00:00Z",
			ContainerStatuses: []ContainerStatus{{
				Name:         "worker",
				RestartCount: 2,
				Ready:        true,
			}},
		},
	}

	identity, snapshot, ok := MapPod(pod, "worker")
	if !ok {
		t.Fatalf("expected pod mapping to succeed")
	}
	if identity.WorkerID != "worker-1" {
		t.Fatalf("worker_id = %q, want worker-1", identity.WorkerID)
	}
	if snapshot.Phase != domain.PodPhaseRunning {
		t.Fatalf("phase = %q, want %q", snapshot.Phase, domain.PodPhaseRunning)
	}
	if snapshot.RestartCount != 2 {
		t.Fatalf("restart_count = %d, want 2", snapshot.RestartCount)
	}
}

func TestClientListPods(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("labelSelector"); got != "app=worker" {
			t.Fatalf("labelSelector = %q, want app=worker", got)
		}
		_ = json.NewEncoder(w).Encode(PodList{
			Items: []Pod{{
				Metadata: PodMetadata{Name: "worker-1", Namespace: "sim", UID: "uid-1"},
			}},
			Metadata: listMeta{ResourceVersion: "123"},
		})
	}))
	defer server.Close()

	client, err := NewClient(Config{
		APIHost:       server.URL,
		Namespace:     "sim",
		LabelSelector: "app=worker",
		HTTPClient:    server.Client(),
	})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	list, err := client.ListPods(context.Background())
	if err != nil {
		t.Fatalf("list pods: %v", err)
	}
	if len(list.Items) != 1 {
		t.Fatalf("len(items) = %d, want 1", len(list.Items))
	}
	if list.Metadata.ResourceVersion != "123" {
		t.Fatalf("resource version = %q, want 123", list.Metadata.ResourceVersion)
	}
}

func TestClientWatchPods(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		flusher, ok := w.(http.Flusher)
		if !ok {
			t.Fatalf("response writer is not flushable")
		}
		events := []WatchEvent{
			{Type: "ADDED", Object: Pod{Metadata: PodMetadata{Name: "worker-1"}}},
			{Type: "MODIFIED", Object: Pod{Metadata: PodMetadata{Name: "worker-2"}}},
		}
		for _, event := range events {
			if err := json.NewEncoder(w).Encode(event); err != nil {
				t.Fatalf("encode event: %v", err)
			}
			flusher.Flush()
		}
	}))
	defer server.Close()

	client, err := NewClient(Config{
		APIHost:    server.URL,
		Namespace:  "sim",
		HTTPClient: server.Client(),
	})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	var names []string
	err = client.WatchPods(context.Background(), "10", func(event WatchEvent) error {
		names = append(names, event.Object.Metadata.Name)
		return nil
	})
	if err != nil {
		t.Fatalf("watch pods: %v", err)
	}
	if len(names) != 2 {
		t.Fatalf("len(names) = %d, want 2", len(names))
	}
}

func TestDefaultNamespaceFallsBackToDefault(t *testing.T) {
	client, err := NewClient(Config{
		APIHost:    "https://cluster.example",
		Namespace:  "",
		HTTPClient: &http.Client{Timeout: time.Second},
	})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	if client.cfg.Namespace == "" {
		t.Fatalf("expected default namespace to be set")
	}
}

func TestInventorySyncSyncOnce(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(PodList{
			Items: []Pod{{
				Metadata: PodMetadata{Name: "worker-1", Namespace: "sim", UID: "uid-1"},
				Spec: PodSpec{
					NodeName: "node-a",
					Containers: []Container{{
						Name:  "worker",
						Image: "worker:v1",
					}},
				},
				Status: PodStatus{
					Phase: "Running",
					ContainerStatuses: []ContainerStatus{{
						Name:         "worker",
						RestartCount: 1,
					}},
				},
			}},
			Metadata: listMeta{ResourceVersion: "55"},
		})
	}))
	defer server.Close()

	client, err := NewClient(Config{
		APIHost:         server.URL,
		Namespace:       "sim",
		WorkerContainer: "worker",
		HTTPClient:      server.Client(),
	})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	snapshots := newSnapshotMemoryStore()
	syncer := NewInventorySync(client, snapshots, "worker", domain.DefaultStatusPolicy())
	resourceVersion, err := syncer.SyncOnce(context.Background())
	if err != nil {
		t.Fatalf("sync once: %v", err)
	}
	if resourceVersion != "55" {
		t.Fatalf("resource version = %q, want 55", resourceVersion)
	}
	snapshot, ok, err := snapshots.GetWorkerSnapshot("worker-1")
	if err != nil {
		t.Fatalf("get worker snapshot: %v", err)
	}
	if !ok {
		t.Fatalf("expected snapshot to exist")
	}
	if snapshot.Identity.NodeName != "node-a" {
		t.Fatalf("node_name = %q, want node-a", snapshot.Identity.NodeName)
	}
	if snapshot.Pod.Phase != domain.PodPhaseRunning {
		t.Fatalf("phase = %q, want %q", snapshot.Pod.Phase, domain.PodPhaseRunning)
	}
}

type snapshotMemoryStore struct {
	snapshots map[domain.WorkerID]domain.WorkerSnapshot
}

func newSnapshotMemoryStore() *snapshotMemoryStore {
	return &snapshotMemoryStore{snapshots: make(map[domain.WorkerID]domain.WorkerSnapshot)}
}

func (s *snapshotMemoryStore) GetWorkerSnapshot(workerID domain.WorkerID) (domain.WorkerSnapshot, bool, error) {
	snapshot, ok := s.snapshots[workerID]
	return snapshot, ok, nil
}

func (s *snapshotMemoryStore) UpsertWorkerSnapshot(snapshot domain.WorkerSnapshot) error {
	s.snapshots[snapshot.Identity.WorkerID] = snapshot
	return nil
}
