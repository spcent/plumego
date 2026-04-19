package kube

import (
	"strings"
	"time"

	"github.com/spcent/plumego/reference/workerfleet/internal/domain"
)

type Pod struct {
	Metadata PodMetadata `json:"metadata"`
	Spec     PodSpec     `json:"spec"`
	Status   PodStatus   `json:"status"`
}

type PodMetadata struct {
	Name              string `json:"name"`
	Namespace         string `json:"namespace"`
	UID               string `json:"uid"`
	ResourceVersion   string `json:"resourceVersion"`
	DeletionTimestamp string `json:"deletionTimestamp,omitempty"`
}

type PodSpec struct {
	NodeName   string      `json:"nodeName"`
	Containers []Container `json:"containers"`
}

type Container struct {
	Name  string `json:"name"`
	Image string `json:"image"`
}

type PodStatus struct {
	Phase             string            `json:"phase"`
	PodIP             string            `json:"podIP"`
	HostIP            string            `json:"hostIP"`
	StartTime         string            `json:"startTime,omitempty"`
	ContainerStatuses []ContainerStatus `json:"containerStatuses,omitempty"`
}

type ContainerStatus struct {
	Name         string `json:"name"`
	Image        string `json:"image,omitempty"`
	RestartCount int32  `json:"restartCount"`
	Ready        bool   `json:"ready"`
}

func MapPod(pod Pod, workerContainer string) (domain.WorkerIdentity, domain.PodSnapshot, bool) {
	identity := domain.WorkerIdentity{
		WorkerID:  domain.WorkerID(strings.TrimSpace(pod.Metadata.Name)),
		Namespace: strings.TrimSpace(pod.Metadata.Namespace),
		PodName:   strings.TrimSpace(pod.Metadata.Name),
		PodUID:    domain.PodUID(strings.TrimSpace(pod.Metadata.UID)),
		NodeName:  strings.TrimSpace(pod.Spec.NodeName),
	}
	if identity.WorkerID == "" || identity.PodName == "" {
		return domain.WorkerIdentity{}, domain.PodSnapshot{}, false
	}

	container, status, ok := selectWorkerContainer(pod, workerContainer)
	if !ok {
		return domain.WorkerIdentity{}, domain.PodSnapshot{}, false
	}

	identity.ContainerName = container.Name
	identity.Image = firstNonEmpty(status.Image, container.Image)

	snapshot := domain.PodSnapshot{
		Phase:        mapPodPhase(pod.Status.Phase),
		PodIP:        strings.TrimSpace(pod.Status.PodIP),
		HostIP:       strings.TrimSpace(pod.Status.HostIP),
		StartedAt:    parseTime(pod.Status.StartTime),
		RestartCount: status.RestartCount,
		DeletedAt:    parseTime(pod.Metadata.DeletionTimestamp),
	}

	return identity, snapshot, true
}

func selectWorkerContainer(pod Pod, workerContainer string) (Container, ContainerStatus, bool) {
	if workerContainer != "" {
		for _, container := range pod.Spec.Containers {
			if container.Name != workerContainer {
				continue
			}
			return container, findContainerStatus(pod.Status.ContainerStatuses, container.Name), true
		}
		return Container{}, ContainerStatus{}, false
	}

	if len(pod.Spec.Containers) == 0 {
		return Container{}, ContainerStatus{}, false
	}
	container := pod.Spec.Containers[0]
	return container, findContainerStatus(pod.Status.ContainerStatuses, container.Name), true
}

func findContainerStatus(statuses []ContainerStatus, name string) ContainerStatus {
	for _, status := range statuses {
		if status.Name == name {
			return status
		}
	}
	return ContainerStatus{Name: name}
}

func mapPodPhase(phase string) domain.PodPhase {
	switch strings.ToLower(strings.TrimSpace(phase)) {
	case "pending":
		return domain.PodPhasePending
	case "running":
		return domain.PodPhaseRunning
	case "succeeded":
		return domain.PodPhaseSucceeded
	case "failed":
		return domain.PodPhaseFailed
	default:
		return domain.PodPhaseUnknown
	}
}

func parseTime(value string) time.Time {
	if strings.TrimSpace(value) == "" {
		return time.Time{}
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}
	}
	return parsed
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
