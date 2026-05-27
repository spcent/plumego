package domain

import "time"

func ReconcilePodSnapshot(snapshot WorkerSnapshot, identity WorkerIdentity, pod PodSnapshot, now time.Time, policy StatusPolicy) (WorkerSnapshot, []DomainEvent) {
	snapshot.Identity = mergeIdentity(snapshot.Identity, identity)
	return MergePodSnapshot(snapshot, pod, now, policy)
}
