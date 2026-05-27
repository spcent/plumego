package mongo

import "workerfleet/internal/domain"

func cloneStringMap(input map[string]string) map[string]string {
	if len(input) == 0 {
		return nil
	}
	out := make(map[string]string, len(input))
	for key, value := range input {
		out[key] = value
	}
	return out
}

func cloneTasks(tasks []domain.ActiveTask) []domain.ActiveTask {
	if len(tasks) == 0 {
		return nil
	}
	out := make([]domain.ActiveTask, 0, len(tasks))
	for _, task := range tasks {
		task.Metadata = cloneStringMap(task.Metadata)
		out = append(out, task)
	}
	return out
}
