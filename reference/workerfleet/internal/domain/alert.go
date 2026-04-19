package domain

type AlertType string

const (
	AlertWorkerOffline      AlertType = "worker_offline"
	AlertWorkerDegraded     AlertType = "worker_degraded"
	AlertWorkerNotAccepting AlertType = "worker_not_accepting_tasks"
	AlertWorkerNoHeartbeat  AlertType = "worker_no_heartbeat"
	AlertWorkerStageStuck   AlertType = "worker_stage_stuck"
	AlertPodRestartBurst    AlertType = "pod_restart_burst"
	AlertPodMissing         AlertType = "pod_missing"
	AlertTaskConflict       AlertType = "task_conflict"
)
