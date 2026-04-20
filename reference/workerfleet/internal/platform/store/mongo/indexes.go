package mongo

import (
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type IndexSpec struct {
	Name               string
	Keys               bson.D
	Unique             bool
	ExpireAfterSeconds *int32
}

func (s IndexSpec) Model() mongo.IndexModel {
	builder := options.Index().SetName(s.Name)
	if s.Unique {
		builder = builder.SetUnique(true)
	}
	if s.ExpireAfterSeconds != nil {
		builder = builder.SetExpireAfterSeconds(*s.ExpireAfterSeconds)
	}
	return mongo.IndexModel{
		Keys:    s.Keys,
		Options: builder,
	}
}

func CollectionIndexSpecs() map[string][]IndexSpec {
	ttlZero := int32(0)

	return map[string][]IndexSpec{
		CollectionWorkerSnapshots: {
			{Name: "worker_status_namespace", Keys: bson.D{{Key: "worker_status", Value: 1}, {Key: "namespace", Value: 1}}},
			{Name: "node_status", Keys: bson.D{{Key: "node_name", Value: 1}, {Key: "worker_status", Value: 1}}},
			{Name: "accepting_status", Keys: bson.D{{Key: "accepting_tasks", Value: 1}, {Key: "worker_status", Value: 1}}},
			{Name: "updated_at_desc", Keys: bson.D{{Key: "updated_at", Value: -1}}},
		},
		CollectionWorkerActive: {
			{Name: "task_id_unique", Keys: bson.D{{Key: "task_id", Value: 1}}, Unique: true},
			{Name: "worker_updated_desc", Keys: bson.D{{Key: "worker_id", Value: 1}, {Key: "updated_at", Value: -1}}},
			{Name: "task_type_phase", Keys: bson.D{{Key: "task_type", Value: 1}, {Key: "phase", Value: 1}}},
		},
		CollectionTaskHistory: {
			{Name: "task_updated_desc", Keys: bson.D{{Key: "task_id", Value: 1}, {Key: "last_updated_at", Value: -1}}},
			{Name: "worker_updated_desc", Keys: bson.D{{Key: "worker_id", Value: 1}, {Key: "last_updated_at", Value: -1}}},
			{Name: "expire_at_ttl", Keys: bson.D{{Key: "expire_at", Value: 1}}, ExpireAfterSeconds: &ttlZero},
		},
		CollectionWorkerEvents: {
			{Name: "worker_occurred_desc", Keys: bson.D{{Key: "worker_id", Value: 1}, {Key: "occurred_at", Value: -1}}},
			{Name: "task_occurred_desc", Keys: bson.D{{Key: "task_id", Value: 1}, {Key: "occurred_at", Value: -1}}},
			{Name: "event_type_occurred_desc", Keys: bson.D{{Key: "event_type", Value: 1}, {Key: "occurred_at", Value: -1}}},
			{Name: "expire_at_ttl", Keys: bson.D{{Key: "expire_at", Value: 1}}, ExpireAfterSeconds: &ttlZero},
		},
		CollectionAlertEvents: {
			{Name: "dedupe_status", Keys: bson.D{{Key: "dedupe_key", Value: 1}, {Key: "status", Value: 1}}},
			{Name: "worker_triggered_desc", Keys: bson.D{{Key: "worker_id", Value: 1}, {Key: "triggered_at", Value: -1}}},
			{Name: "alert_type_status", Keys: bson.D{{Key: "alert_type", Value: 1}, {Key: "status", Value: 1}}},
			{Name: "expire_at_ttl", Keys: bson.D{{Key: "expire_at", Value: 1}}, ExpireAfterSeconds: &ttlZero},
		},
	}
}
