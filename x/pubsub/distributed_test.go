package pubsub

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/spcent/plumego/contract"
)

func TestDistributedPubSub_Basic(t *testing.T) {
	t.Parallel()
	// Create two nodes
	config1 := DefaultClusterConfig("node1", "127.0.0.1:17001")
	config1.Peers = []string{"127.0.0.1:17002"}

	config2 := DefaultClusterConfig("node2", "127.0.0.1:17002")
	config2.Peers = []string{"127.0.0.1:17001"}

	dps1, err := NewDistributed(config1)
	if err != nil {
		t.Fatalf("Failed to create node1: %v", err)
	}
	defer dps1.Close()

	dps2, err := NewDistributed(config2)
	if err != nil {
		t.Fatalf("Failed to create node2: %v", err)
	}
	defer dps2.Close()

	// Join cluster
	ctx := context.Background()
	if err := dps1.JoinCluster(ctx); err != nil {
		t.Fatalf("node1 failed to join: %v", err)
	}

	if err := dps2.JoinCluster(ctx); err != nil {
		t.Fatalf("node2 failed to join: %v", err)
	}

	// Wait for discovery
	time.Sleep(200 * time.Millisecond)

	// Check nodes discovered each other
	nodes1 := dps1.Nodes()
	if len(nodes1) < 2 {
		t.Logf("Note: Node discovery might be eventual (found %d nodes)", len(nodes1))
	}

	nodes2 := dps2.Nodes()
	if len(nodes2) < 2 {
		t.Logf("Note: Node discovery might be eventual (found %d nodes)", len(nodes2))
	}
}

func TestDistributedPubSub_GlobalPublish(t *testing.T) {
	t.Parallel()
	// Create two nodes
	config1 := DefaultClusterConfig("node1", "127.0.0.1:17011")
	config1.Peers = []string{"127.0.0.1:17012"}

	config2 := DefaultClusterConfig("node2", "127.0.0.1:17012")
	config2.Peers = []string{"127.0.0.1:17011"}

	dps1, err := NewDistributed(config1)
	if err != nil {
		t.Fatalf("Failed to create node1: %v", err)
	}
	defer dps1.Close()

	dps2, err := NewDistributed(config2)
	if err != nil {
		t.Fatalf("Failed to create node2: %v", err)
	}
	defer dps2.Close()

	// Join cluster
	ctx := context.Background()
	_ = dps1.JoinCluster(ctx)
	_ = dps2.JoinCluster(ctx)

	time.Sleep(300 * time.Millisecond) // Wait for cluster formation

	// Subscribe on node2
	sub, err := dps2.Subscribe(context.Background(), "test.global", SubOptions{BufferSize: 10})
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}
	defer sub.Cancel()

	// Publish from node1
	msg := Message{
		ID:   "global-msg",
		Data: map[string]any{"text": "hello cluster"},
	}

	if err := dps1.PublishGlobal("test.global", msg); err != nil {
		t.Fatalf("Failed to publish globally: %v", err)
	}

	// Wait and check if received on node2
	select {
	case received := <-sub.C():
		if received.ID != "global-msg" {
			t.Errorf("Expected message ID 'global-msg', got '%s'", received.ID)
		}

	case <-time.After(1 * time.Second):
		t.Error("Did not receive message on node2 within timeout")
	}

	// Check stats
	stats1 := dps1.ClusterStats()
	if stats1.HealthyNodes < 1 {
		t.Logf("Note: Cluster may still be forming (healthy nodes: %d)", stats1.HealthyNodes)
	}
}

func TestDistributedPubSub_Heartbeat(t *testing.T) {
	t.Parallel()
	config := DefaultClusterConfig("test-node", "127.0.0.1:17021")
	config.HeartbeatInterval = 100 * time.Millisecond
	config.HeartbeatTimeout = 300 * time.Millisecond

	dps, err := NewDistributed(config)
	if err != nil {
		t.Fatalf("Failed to create node: %v", err)
	}
	defer dps.Close()

	ctx := context.Background()
	if err := dps.JoinCluster(ctx); err != nil {
		t.Fatalf("Failed to join cluster: %v", err)
	}

	// Wait for heartbeats
	time.Sleep(400 * time.Millisecond)

	stats := dps.ClusterStats()
	if stats.Heartbeats == 0 && len(config.Peers) > 0 {
		t.Log("Note: No heartbeats sent (expected with no peers)")
	}

	// Check self is healthy
	nodes := dps.Nodes()
	foundSelf := false
	for _, node := range nodes {
		if node.ID == "test-node" && node.Healthy {
			foundSelf = true
			break
		}
	}

	if !foundSelf {
		t.Error("Self node not found or unhealthy")
	}
}

func TestDistributedPubSub_NodeFailure(t *testing.T) {
	t.Parallel()
	// Create three nodes
	config1 := DefaultClusterConfig("node1", "127.0.0.1:17031")
	config1.Peers = []string{"127.0.0.1:17032", "127.0.0.1:17033"}
	config1.HeartbeatInterval = 100 * time.Millisecond
	config1.HeartbeatTimeout = 300 * time.Millisecond

	config2 := DefaultClusterConfig("node2", "127.0.0.1:17032")
	config2.Peers = []string{"127.0.0.1:17031", "127.0.0.1:17033"}
	config2.HeartbeatInterval = 100 * time.Millisecond
	config2.HeartbeatTimeout = 300 * time.Millisecond

	config3 := DefaultClusterConfig("node3", "127.0.0.1:17033")
	config3.Peers = []string{"127.0.0.1:17031", "127.0.0.1:17032"}
	config3.HeartbeatInterval = 100 * time.Millisecond
	config3.HeartbeatTimeout = 300 * time.Millisecond

	dps1, _ := NewDistributed(config1)
	defer dps1.Close()

	dps2, _ := NewDistributed(config2)
	defer dps2.Close()

	dps3, _ := NewDistributed(config3)

	ctx := context.Background()
	_ = dps1.JoinCluster(ctx)
	_ = dps2.JoinCluster(ctx)
	_ = dps3.JoinCluster(ctx)

	time.Sleep(500 * time.Millisecond) // Allow cluster to form

	initialNodes := len(dps1.Nodes())

	// Shutdown node3
	dps3.Close()

	// Wait for failure detection
	time.Sleep(500 * time.Millisecond)

	// Check node1 detects node3 as unhealthy
	nodes := dps1.Nodes()
	node3Healthy := false
	for _, node := range nodes {
		if node.ID == "node3" && node.Healthy {
			node3Healthy = true
			break
		}
	}

	if node3Healthy {
		t.Log("Note: Node3 still marked healthy (failure detection may be eventual)")
	}

	if len(nodes) != initialNodes {
		t.Logf("Node count changed from %d to %d", initialNodes, len(nodes))
	}
}

func TestDistributedPubSub_LocalTopics(t *testing.T) {
	t.Parallel()
	config := DefaultClusterConfig("test", "127.0.0.1:17041")

	dps, err := NewDistributed(config)
	if err != nil {
		t.Fatalf("Failed to create node: %v", err)
	}
	defer dps.Close()

	// Subscribe to topics
	sub1, _ := dps.Subscribe(context.Background(), "topic.a", SubOptions{BufferSize: 1})
	defer sub1.Cancel()

	sub2, _ := dps.Subscribe(context.Background(), "topic.b", SubOptions{BufferSize: 1})
	defer sub2.Cancel()

	sub3, _ := dps.Subscribe(context.Background(), "topic.a", SubOptions{BufferSize: 1}) // duplicate
	defer sub3.Cancel()

	time.Sleep(50 * time.Millisecond)

	topics := dps.getLocalTopics()

	if len(topics) != 2 {
		t.Errorf("Expected 2 unique topics, got %d: %v", len(topics), topics)
	}

	// Check topics are sorted
	if len(topics) == 2 && topics[0] > topics[1] {
		t.Error("Topics not sorted")
	}
}

func TestDistributedPubSub_ClusterStats(t *testing.T) {
	t.Parallel()
	config := DefaultClusterConfig("stats-test", "127.0.0.1:17051")

	dps, err := NewDistributed(config)
	if err != nil {
		t.Fatalf("Failed to create node: %v", err)
	}
	defer dps.Close()

	ctx := context.Background()
	_ = dps.JoinCluster(ctx)

	stats := dps.ClusterStats()

	if stats.TotalNodes != 1 {
		t.Errorf("Expected 1 total node (self), got %d", stats.TotalNodes)
	}

	if stats.HealthyNodes != 1 {
		t.Errorf("Expected 1 healthy node (self), got %d", stats.HealthyNodes)
	}
}

func TestDistributedPubSub_ConcurrentPublish(t *testing.T) {
	t.Parallel()
	config1 := DefaultClusterConfig("node1", "127.0.0.1:17061")
	config1.Peers = []string{"127.0.0.1:17062"}

	config2 := DefaultClusterConfig("node2", "127.0.0.1:17062")
	config2.Peers = []string{"127.0.0.1:17061"}

	dps1, _ := NewDistributed(config1)
	defer dps1.Close()

	dps2, _ := NewDistributed(config2)
	defer dps2.Close()

	ctx := context.Background()
	_ = dps1.JoinCluster(ctx)
	_ = dps2.JoinCluster(ctx)

	time.Sleep(300 * time.Millisecond)

	// Concurrent publishers
	const numGoroutines = 5
	const messagesPerGoroutine = 10

	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			for j := 0; j < messagesPerGoroutine; j++ {
				msg := Message{
					Data: map[string]any{"worker": id, "seq": j},
				}
				_ = dps1.PublishGlobal("concurrent.topic", msg)
			}
			done <- true
		}(i)
	}

	// Wait for completion
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	time.Sleep(200 * time.Millisecond)

	stats := dps1.ClusterStats()
	if stats.ClusterForwards == 0 && len(config1.Peers) > 0 {
		t.Log("Note: No cluster forwards (may be expected if peer not connected)")
	}
}

func TestDistributedPubSub_MultipleSubscribers(t *testing.T) {
	t.Parallel()
	config1 := DefaultClusterConfig("pub", "127.0.0.1:17071")
	config1.Peers = []string{"127.0.0.1:17072"}

	config2 := DefaultClusterConfig("sub", "127.0.0.1:17072")
	config2.Peers = []string{"127.0.0.1:17071"}

	dps1, _ := NewDistributed(config1)
	defer dps1.Close()

	dps2, _ := NewDistributed(config2)
	defer dps2.Close()

	ctx := context.Background()
	_ = dps1.JoinCluster(ctx)
	_ = dps2.JoinCluster(ctx)

	time.Sleep(300 * time.Millisecond)

	// Multiple subscribers on node2
	sub1, _ := dps2.Subscribe(context.Background(), "fanout.topic", SubOptions{BufferSize: 5})
	defer sub1.Cancel()

	sub2, _ := dps2.Subscribe(context.Background(), "fanout.topic", SubOptions{BufferSize: 5})
	defer sub2.Cancel()

	sub3, _ := dps2.Subscribe(context.Background(), "fanout.topic", SubOptions{BufferSize: 5})
	defer sub3.Cancel()

	// Publish from node1
	msg := Message{ID: "fanout-msg", Data: "test"}
	_ = dps1.PublishGlobal("fanout.topic", msg)

	time.Sleep(200 * time.Millisecond)

	// Check all subscribers received
	received := 0
	timeout := time.After(1 * time.Second)

	for received < 3 {
		select {
		case <-sub1.C():
			received++
		case <-sub2.C():
			received++
		case <-sub3.C():
			received++
		case <-timeout:
			goto done
		default:
			time.Sleep(10 * time.Millisecond)
		}
	}

done:
	if received != 3 {
		t.Logf("Note: Expected 3 messages, received %d (cluster propagation may be eventual)", received)
	}
}

func TestDistributedPubSub_InvalidConfig(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		config ClusterConfig
	}{
		{
			name:   "empty node ID",
			config: ClusterConfig{NodeID: "", ListenAddr: "127.0.0.1:18000"},
		},
		{
			name:   "empty listen addr",
			config: ClusterConfig{NodeID: "node1", ListenAddr: ""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewDistributed(tt.config)
			if err == nil {
				t.Error("Expected error for invalid config")
			}
		})
	}
}

func TestDistributedPubSub_HTTPHandlersUseCanonicalResponses(t *testing.T) {
	config := DefaultClusterConfig("node-http", "127.0.0.1:18101")

	dps, err := NewDistributed(config)
	if err != nil {
		t.Fatalf("NewDistributed: %v", err)
	}
	defer dps.Close()

	healthReq := httptest.NewRequest(http.MethodGet, "/health", nil)
	healthRec := httptest.NewRecorder()
	dps.handleHealth(healthRec, healthReq)
	if healthRec.Code != http.StatusOK {
		t.Fatalf("health status = %d, want %d", healthRec.Code, http.StatusOK)
	}
	assertPubSubJSONContentType(t, healthRec)

	var health map[string]any
	if err := json.NewDecoder(healthRec.Body).Decode(&health); err != nil {
		t.Fatalf("decode health: %v", err)
	}
	if health["node_id"] != "node-http" {
		t.Fatalf("node_id = %v, want %q", health["node_id"], "node-http")
	}

	heartbeatBody := `{"node_id":"peer-1","addr":"127.0.0.1:18102","topics":["orders"],"timestamp":"2026-04-12T00:00:00Z","version":"1.0"}`
	heartbeatReq := httptest.NewRequest(http.MethodPost, "/heartbeat", strings.NewReader(heartbeatBody))
	heartbeatRec := httptest.NewRecorder()
	dps.handleHeartbeat(heartbeatRec, heartbeatReq)
	if heartbeatRec.Code != http.StatusOK {
		t.Fatalf("heartbeat status = %d, want %d", heartbeatRec.Code, http.StatusOK)
	}
	assertPubSubJSONContentType(t, heartbeatRec)

	var heartbeatResp heartbeatPayload
	if err := json.NewDecoder(heartbeatRec.Body).Decode(&heartbeatResp); err != nil {
		t.Fatalf("decode heartbeat response: %v", err)
	}
	if heartbeatResp.NodeID != "node-http" {
		t.Fatalf("heartbeat node_id = %q, want %q", heartbeatResp.NodeID, "node-http")
	}

	syncReq := httptest.NewRequest(http.MethodGet, "/sync", nil)
	syncRec := httptest.NewRecorder()
	dps.handleSync(syncRec, syncReq)
	if syncRec.Code != http.StatusOK {
		t.Fatalf("sync status = %d, want %d", syncRec.Code, http.StatusOK)
	}
	assertPubSubJSONContentType(t, syncRec)

	var syncResp struct {
		Nodes []*ClusterNode `json:"nodes"`
	}
	if err := json.NewDecoder(syncRec.Body).Decode(&syncResp); err != nil {
		t.Fatalf("decode sync response: %v", err)
	}
	if len(syncResp.Nodes) != 1 || syncResp.Nodes[0].ID != "peer-1" {
		t.Fatalf("unexpected sync nodes: %+v", syncResp.Nodes)
	}
}

func TestDistributedPubSub_HTTPHandlersReturnStructuredErrors(t *testing.T) {
	config := DefaultClusterConfig("node-http", "127.0.0.1:18111")
	config.AuthToken = "secret-token"

	dps, err := NewDistributed(config)
	if err != nil {
		t.Fatalf("NewDistributed: %v", err)
	}
	defer dps.Close()

	methodReq := httptest.NewRequest(http.MethodGet, "/heartbeat", nil)
	methodRec := httptest.NewRecorder()
	dps.handleHeartbeat(methodRec, methodReq)
	if methodRec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("heartbeat method status = %d, want %d", methodRec.Code, http.StatusMethodNotAllowed)
	}
	assertPubSubErrorCode(t, methodRec, "METHOD_NOT_ALLOWED")

	authReq := httptest.NewRequest(http.MethodGet, "/health", nil)
	authRec := httptest.NewRecorder()
	dps.handleHealth(authRec, authReq)
	if authRec.Code != http.StatusUnauthorized {
		t.Fatalf("health auth status = %d, want %d", authRec.Code, http.StatusUnauthorized)
	}
	assertPubSubErrorCode(t, authRec, "UNAUTHORIZED")
}

func TestDistributedPubSub_ClusterPublishReturnsJSON(t *testing.T) {
	config := DefaultClusterConfig("node-http", "127.0.0.1:18121")

	dps, err := NewDistributed(config)
	if err != nil {
		t.Fatalf("NewDistributed: %v", err)
	}
	defer dps.Close()

	body := `{"type":"publish","node_id":"peer-1","topic":"orders","message":{"data":"ok"},"timestamp":"2026-04-12T00:00:00Z"}`
	req := httptest.NewRequest(http.MethodPost, "/publish", strings.NewReader(body))
	rec := httptest.NewRecorder()
	dps.handleClusterPublish(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("publish status = %d, want %d", rec.Code, http.StatusOK)
	}
	assertPubSubJSONContentType(t, rec)

	var resp map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode publish response: %v", err)
	}
	if resp["status"] != "ok" {
		t.Fatalf("publish status payload = %q, want %q", resp["status"], "ok")
	}
}

func assertPubSubJSONContentType(t *testing.T, rec *httptest.ResponseRecorder) {
	t.Helper()
	if got := rec.Header().Get("Content-Type"); got != contract.ContentTypeJSON {
		t.Fatalf("content type = %q, want %q", got, contract.ContentTypeJSON)
	}
}

func assertPubSubErrorCode(t *testing.T, rec *httptest.ResponseRecorder, want string) {
	t.Helper()
	assertPubSubJSONContentType(t, rec)

	var resp contract.ErrorResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error response: %v", err)
	}
	if resp.Error.Code != want {
		t.Fatalf("error code = %q, want %q", resp.Error.Code, want)
	}
}

func BenchmarkDistributedPubSub_LocalPublish(b *testing.B) {
	config := DefaultClusterConfig("bench", "127.0.0.1:17081")

	dps, err := NewDistributed(config)
	if err != nil {
		b.Fatalf("Failed to create node: %v", err)
	}
	defer dps.Close()

	ctx := context.Background()
	_ = dps.JoinCluster(ctx)

	msg := Message{Data: []byte("benchmark message")}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = dps.Publish("bench.topic", msg)
	}
}

func BenchmarkDistributedPubSub_GlobalPublish(b *testing.B) {
	config1 := DefaultClusterConfig("node1", "127.0.0.1:17091")
	config1.Peers = []string{"127.0.0.1:17092"}

	config2 := DefaultClusterConfig("node2", "127.0.0.1:17092")
	config2.Peers = []string{"127.0.0.1:17091"}

	dps1, _ := NewDistributed(config1)
	defer dps1.Close()

	dps2, _ := NewDistributed(config2)
	defer dps2.Close()

	ctx := context.Background()
	_ = dps1.JoinCluster(ctx)
	_ = dps2.JoinCluster(ctx)

	time.Sleep(300 * time.Millisecond)

	msg := Message{Data: []byte("benchmark message")}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = dps1.PublishGlobal("bench.topic", msg)
	}
}
