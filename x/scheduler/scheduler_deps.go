package scheduler

func (s *Scheduler) wouldCreateDependencyCycleLocked(newID JobID, dependencies []JobID) bool {
	if len(dependencies) == 0 {
		return false
	}
	for _, depID := range dependencies {
		if depID == newID {
			return true
		}
		visited := map[JobID]bool{}
		if s.depPathExistsLocked(depID, newID, visited) {
			return true
		}
	}
	return false
}

func (s *Scheduler) depPathExistsLocked(current, target JobID, visited map[JobID]bool) bool {
	if current == target {
		return true
	}
	if visited[current] {
		return false
	}
	visited[current] = true
	j, ok := s.jobs[current]
	if !ok {
		return false
	}
	for _, depID := range j.options.Dependencies {
		if s.depPathExistsLocked(depID, target, visited) {
			return true
		}
	}
	return false
}
