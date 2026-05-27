package mongo

import (
	"reflect"
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"
	"workerfleet/internal/domain"
	platformstore "workerfleet/internal/platform/store"
)

func TestWorkerSnapshotFilterDocUsesScalarEquality(t *testing.T) {
	accepting := true
	got := workerSnapshotFilterDoc(platformstore.WorkerSnapshotFilter{
		Status:         domain.WorkerStatusOnline,
		Namespace:      "sim",
		NodeName:       "node-a",
		TaskType:       "simulation",
		AcceptingTasks: &accepting,
	})
	want := bson.D{
		{Key: "worker_status", Value: bson.D{{Key: "$eq", Value: "online"}}},
		{Key: "namespace", Value: bson.D{{Key: "$eq", Value: "sim"}}},
		{Key: "node_name", Value: bson.D{{Key: "$eq", Value: "node-a"}}},
		{Key: "accepting_tasks", Value: bson.D{{Key: "$eq", Value: true}}},
		{Key: "active_tasks.task_type", Value: bson.D{{Key: "$eq", Value: "simulation"}}},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("filter = %#v, want %#v", got, want)
	}
}

func TestAlertFilterDocUsesScalarEquality(t *testing.T) {
	got := alertFilterDoc(platformstore.AlertFilter{
		WorkerID:  "worker-1",
		AlertType: domain.AlertWorkerOffline,
		Status:    domain.AlertStatusResolved,
	})
	want := bson.D{
		{Key: "worker_id", Value: bson.D{{Key: "$eq", Value: "worker-1"}}},
		{Key: "alert_type", Value: bson.D{{Key: "$eq", Value: "worker_offline"}}},
		{Key: "status", Value: bson.D{{Key: "$eq", Value: "resolved"}}},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("filter = %#v, want %#v", got, want)
	}
}

func TestMongoIDFiltersUseScalarEquality(t *testing.T) {
	tests := []struct {
		name string
		got  bson.D
		want bson.D
	}{
		{
			name: "snapshot worker id",
			got:  workerSnapshotIDFilter("worker-1"),
			want: bson.D{{Key: "_id", Value: bson.D{{Key: "$eq", Value: "worker-1"}}}},
		},
		{
			name: "active worker id",
			got:  workerActiveByWorkerIDFilter("worker-1"),
			want: bson.D{{Key: "worker_id", Value: bson.D{{Key: "$eq", Value: "worker-1"}}}},
		},
		{
			name: "active task id",
			got:  workerActiveByTaskIDFilter("task-1"),
			want: bson.D{{Key: "task_id", Value: bson.D{{Key: "$eq", Value: "task-1"}}}},
		},
		{
			name: "history task id",
			got:  taskHistoryByTaskIDFilter("task-1"),
			want: bson.D{{Key: "task_id", Value: bson.D{{Key: "$eq", Value: "task-1"}}}},
		},
		{
			name: "event worker id",
			got:  workerEventFilterDoc("worker-1"),
			want: bson.D{{Key: "worker_id", Value: bson.D{{Key: "$eq", Value: "worker-1"}}}},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if !reflect.DeepEqual(test.got, test.want) {
				t.Fatalf("filter = %#v, want %#v", test.got, test.want)
			}
		})
	}
}
