package mongo

import (
	"go.mongodb.org/mongo-driver/v2/bson"
	"workerfleet/internal/domain"
	platformstore "workerfleet/internal/platform/store"
)

func workerSnapshotIDFilter(workerID domain.WorkerID) bson.D {
	return bson.D{stringEq("_id", string(workerID))}
}

func workerActiveByWorkerIDFilter(workerID domain.WorkerID) bson.D {
	return bson.D{stringEq("worker_id", string(workerID))}
}

func workerActiveByTaskIDFilter(taskID domain.TaskID) bson.D {
	return bson.D{stringEq("task_id", string(taskID))}
}

func taskHistoryByTaskIDFilter(taskID domain.TaskID) bson.D {
	return bson.D{stringEq("task_id", string(taskID))}
}

func workerEventFilterDoc(workerID domain.WorkerID) bson.D {
	filter := bson.D{}
	if workerID != "" {
		filter = append(filter, stringEq("worker_id", string(workerID)))
	}
	return filter
}

func workerSnapshotFilterDoc(filter platformstore.WorkerSnapshotFilter) bson.D {
	doc := bson.D{}
	if filter.Status != "" {
		doc = append(doc, stringEq("worker_status", string(filter.Status)))
	}
	if filter.Namespace != "" {
		doc = append(doc, stringEq("namespace", filter.Namespace))
	}
	if filter.NodeName != "" {
		doc = append(doc, stringEq("node_name", filter.NodeName))
	}
	if filter.AcceptingTasks != nil {
		doc = append(doc, boolEq("accepting_tasks", *filter.AcceptingTasks))
	}
	if filter.TaskType != "" {
		doc = append(doc, stringEq("active_tasks.task_type", filter.TaskType))
	}
	return doc
}

func alertFilterDoc(filter platformstore.AlertFilter) bson.D {
	doc := bson.D{}
	if filter.WorkerID != "" {
		doc = append(doc, stringEq("worker_id", string(filter.WorkerID)))
	}
	if filter.AlertType != "" {
		doc = append(doc, stringEq("alert_type", string(filter.AlertType)))
	}
	if filter.Status != "" {
		doc = append(doc, stringEq("status", string(filter.Status)))
	}
	return doc
}

// stringEq keeps caller-provided strings as equality operands under fixed fields.
func stringEq(field string, value string) bson.E {
	return bson.E{Key: field, Value: bson.D{{Key: "$eq", Value: value}}}
}

func boolEq(field string, value bool) bson.E {
	return bson.E{Key: field, Value: bson.D{{Key: "$eq", Value: value}}}
}
