package handler

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/router"
	workerapp "workerfleet/internal/app"
	"workerfleet/internal/domain"
)

type TaskView struct {
	TaskID      string            `json:"task_id"`
	ExecPlanID  string            `json:"exec_plan_id,omitempty"`
	TaskType    string            `json:"task_type,omitempty"`
	Phase       string            `json:"phase,omitempty"`
	PhaseName   string            `json:"phase_name,omitempty"`
	CurrentStep *StepView         `json:"current_step,omitempty"`
	StartedAt   time.Time         `json:"started_at,omitempty"`
	UpdatedAt   time.Time         `json:"updated_at,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

type StepView struct {
	Step       string                `json:"step,omitempty"`
	StepName   string                `json:"step_name,omitempty"`
	Status     domain.CaseStepStatus `json:"status,omitempty"`
	StartedAt  time.Time             `json:"started_at,omitempty"`
	UpdatedAt  time.Time             `json:"updated_at,omitempty"`
	FinishedAt time.Time             `json:"finished_at,omitempty"`
	Attempt    int                   `json:"attempt,omitempty"`
	ErrorClass string                `json:"error_class,omitempty"`
}

type WorkerView struct {
	WorkerID        string     `json:"worker_id"`
	Namespace       string     `json:"namespace,omitempty"`
	PodName         string     `json:"pod_name,omitempty"`
	NodeName        string     `json:"node_name,omitempty"`
	ContainerName   string     `json:"container_name,omitempty"`
	Image           string     `json:"image,omitempty"`
	Version         string     `json:"version,omitempty"`
	Status          string     `json:"status"`
	StatusReason    string     `json:"status_reason,omitempty"`
	ProcessAlive    bool       `json:"process_alive"`
	AcceptingTasks  bool       `json:"accepting_tasks"`
	LastSeenAt      time.Time  `json:"last_seen_at,omitempty"`
	LastReadyAt     time.Time  `json:"last_ready_at,omitempty"`
	ActiveTaskCount int        `json:"active_task_count"`
	ActiveTasks     []TaskView `json:"active_tasks,omitempty"`
}

type WorkerDetail = WorkerView

type WorkerListQuery struct {
	Status         domain.WorkerStatus
	Namespace      string
	NodeName       string
	TaskType       string
	AcceptingTasks *bool
	Page           int
	PageSize       int
}

type WorkerListResult struct {
	Items    []WorkerView `json:"items"`
	Page     int          `json:"page"`
	PageSize int          `json:"page_size"`
	Total    int          `json:"total"`
}

type TaskDetail struct {
	TaskID      string            `json:"task_id"`
	WorkerID    string            `json:"worker_id,omitempty"`
	ExecPlanID  string            `json:"exec_plan_id,omitempty"`
	TaskType    string            `json:"task_type,omitempty"`
	Phase       string            `json:"phase,omitempty"`
	PhaseName   string            `json:"phase_name,omitempty"`
	CurrentStep *StepView         `json:"current_step,omitempty"`
	Status      string            `json:"status,omitempty"`
	StartedAt   time.Time         `json:"started_at,omitempty"`
	UpdatedAt   time.Time         `json:"updated_at,omitempty"`
	EndedAt     time.Time         `json:"ended_at,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

type CaseStepView struct {
	TaskID     string                `json:"task_id"`
	WorkerID   string                `json:"worker_id,omitempty"`
	ExecPlanID string                `json:"exec_plan_id,omitempty"`
	Namespace  string                `json:"namespace,omitempty"`
	PodName    string                `json:"pod_name,omitempty"`
	NodeName   string                `json:"node_name,omitempty"`
	Step       string                `json:"step"`
	StepName   string                `json:"step_name,omitempty"`
	Status     domain.CaseStepStatus `json:"status,omitempty"`
	Result     string                `json:"result,omitempty"`
	ErrorClass string                `json:"error_class,omitempty"`
	Attempt    int                   `json:"attempt,omitempty"`
	StartedAt  time.Time             `json:"started_at,omitempty"`
	FinishedAt time.Time             `json:"finished_at,omitempty"`
	ObservedAt time.Time             `json:"observed_at,omitempty"`
	EventType  domain.EventType      `json:"event_type,omitempty"`
}

type CaseTimelineResult struct {
	TaskID string         `json:"task_id"`
	Items  []CaseStepView `json:"items"`
}

type ExecPlanCaseDrilldownQuery struct {
	ExecPlanID string
	NodeName   string
	PodName    string
	Step       string
	Page       int
	PageSize   int
}

type ExecPlanCaseDrilldownResult struct {
	ExecPlanID string         `json:"exec_plan_id"`
	Items      []CaseStepView `json:"items"`
	Page       int            `json:"page"`
	PageSize   int            `json:"page_size"`
	Total      int            `json:"total"`
}

type FleetSummary struct {
	TotalWorkers     int `json:"total_workers"`
	OnlineWorkers    int `json:"online_workers"`
	DegradedWorkers  int `json:"degraded_workers"`
	OfflineWorkers   int `json:"offline_workers"`
	UnknownWorkers   int `json:"unknown_workers"`
	AcceptingWorkers int `json:"accepting_workers"`
	BusyWorkers      int `json:"busy_workers"`
	ActiveTaskCount  int `json:"active_task_count"`
}

type AlertListQuery struct {
	WorkerID  string
	AlertType string
	Status    string
	Page      int
	PageSize  int
}

type AlertView struct {
	AlertID     string    `json:"alert_id"`
	WorkerID    string    `json:"worker_id,omitempty"`
	TaskID      string    `json:"task_id,omitempty"`
	AlertType   string    `json:"alert_type"`
	Status      string    `json:"status"`
	Severity    string    `json:"severity,omitempty"`
	Message     string    `json:"message"`
	TriggeredAt time.Time `json:"triggered_at"`
	ResolvedAt  time.Time `json:"resolved_at,omitempty"`
}

type AlertListResult struct {
	Items    []AlertView `json:"items"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
	Total    int         `json:"total"`
}

func (h *Handler) ListWorkers(w http.ResponseWriter, r *http.Request) {
	if h.service == nil {
		writeNotImplemented(w, r, "LIST_WORKERS_NOT_CONFIGURED", "list workers service not configured")
		return
	}

	query, ok := parseWorkerListQuery(w, r)
	if !ok {
		return
	}

	result, err := h.service.ListWorkers(r.Context(), query)
	if err != nil {
		writeServiceError(w, r, err)
		return
	}
	_ = contract.WriteResponse(w, r, http.StatusOK, workerListResultFromApp(result), nil)
}

func (h *Handler) GetWorker(w http.ResponseWriter, r *http.Request) {
	if h.service == nil {
		writeNotImplemented(w, r, "GET_WORKER_NOT_CONFIGURED", "get worker service not configured")
		return
	}

	workerID := strings.TrimSpace(router.Param(r, "worker_id"))
	if workerID == "" {
		writeRequiredPathParam(w, r, "worker_id")
		return
	}

	result, err := h.service.GetWorker(r.Context(), domain.WorkerID(workerID))
	if err != nil {
		writeServiceError(w, r, err)
		return
	}
	_ = contract.WriteResponse(w, r, http.StatusOK, workerViewFromApp(result), nil)
}

func (h *Handler) GetTask(w http.ResponseWriter, r *http.Request) {
	if h.service == nil {
		writeNotImplemented(w, r, "GET_TASK_NOT_CONFIGURED", "get task service not configured")
		return
	}

	taskID := strings.TrimSpace(router.Param(r, "task_id"))
	if taskID == "" {
		writeRequiredPathParam(w, r, "task_id")
		return
	}

	result, err := h.service.GetTask(r.Context(), domain.TaskID(taskID))
	if err != nil {
		writeServiceError(w, r, err)
		return
	}
	_ = contract.WriteResponse(w, r, http.StatusOK, taskDetailFromApp(result), nil)
}

func (h *Handler) GetCaseTimeline(w http.ResponseWriter, r *http.Request) {
	if h.service == nil {
		writeNotImplemented(w, r, "GET_CASE_TIMELINE_NOT_CONFIGURED", "get case timeline service not configured")
		return
	}

	taskID := strings.TrimSpace(router.Param(r, "task_id"))
	if taskID == "" {
		writeRequiredPathParam(w, r, "task_id")
		return
	}

	result, err := h.service.GetCaseTimeline(r.Context(), domain.TaskID(taskID))
	if err != nil {
		writeServiceError(w, r, err)
		return
	}
	_ = contract.WriteResponse(w, r, http.StatusOK, caseTimelineResultFromApp(result), nil)
}

func (h *Handler) ListExecPlanCases(w http.ResponseWriter, r *http.Request) {
	if h.service == nil {
		writeNotImplemented(w, r, "LIST_EXEC_PLAN_CASES_NOT_CONFIGURED", "list exec plan cases service not configured")
		return
	}

	execPlanID := strings.TrimSpace(router.Param(r, "exec_plan_id"))
	if execPlanID == "" {
		writeRequiredPathParam(w, r, "exec_plan_id")
		return
	}
	page, pageSize, ok := parsePagination(w, r)
	if !ok {
		return
	}

	result, err := h.service.ListExecPlanCases(r.Context(), workerapp.ExecPlanCaseDrilldownQuery{
		ExecPlanID: execPlanID,
		NodeName:   strings.TrimSpace(r.URL.Query().Get("node_name")),
		PodName:    strings.TrimSpace(r.URL.Query().Get("pod_name")),
		Step:       strings.TrimSpace(r.URL.Query().Get("step")),
		Page:       page,
		PageSize:   pageSize,
	})
	if err != nil {
		writeServiceError(w, r, err)
		return
	}
	_ = contract.WriteResponse(w, r, http.StatusOK, execPlanCaseDrilldownResultFromApp(result), nil)
}

func (h *Handler) FleetSummary(w http.ResponseWriter, r *http.Request) {
	if h.service == nil {
		writeNotImplemented(w, r, "FLEET_SUMMARY_NOT_CONFIGURED", "fleet summary service not configured")
		return
	}

	result, err := h.service.FleetSummary(r.Context())
	if err != nil {
		writeServiceError(w, r, err)
		return
	}
	_ = contract.WriteResponse(w, r, http.StatusOK, fleetSummaryFromApp(result), nil)
}

func (h *Handler) ListAlerts(w http.ResponseWriter, r *http.Request) {
	if h.service == nil {
		writeNotImplemented(w, r, "LIST_ALERTS_NOT_CONFIGURED", "list alerts service not configured")
		return
	}

	page, pageSize, ok := parsePagination(w, r)
	if !ok {
		return
	}

	result, err := h.service.ListAlerts(r.Context(), workerapp.AlertListQuery{
		WorkerID:  strings.TrimSpace(r.URL.Query().Get("worker_id")),
		AlertType: strings.TrimSpace(r.URL.Query().Get("alert_type")),
		Status:    strings.TrimSpace(r.URL.Query().Get("status")),
		Page:      page,
		PageSize:  pageSize,
	})
	if err != nil {
		writeServiceError(w, r, err)
		return
	}
	_ = contract.WriteResponse(w, r, http.StatusOK, alertListResultFromApp(result), nil)
}

func parseWorkerListQuery(w http.ResponseWriter, r *http.Request) (workerapp.WorkerListQuery, bool) {
	page, pageSize, ok := parsePagination(w, r)
	if !ok {
		return workerapp.WorkerListQuery{}, false
	}

	query := workerapp.WorkerListQuery{
		Status:    domain.WorkerStatus(strings.TrimSpace(r.URL.Query().Get("status"))),
		Namespace: strings.TrimSpace(r.URL.Query().Get("namespace")),
		NodeName:  strings.TrimSpace(r.URL.Query().Get("node_name")),
		TaskType:  strings.TrimSpace(r.URL.Query().Get("task_type")),
		Page:      page,
		PageSize:  pageSize,
	}

	rawAccepting := strings.TrimSpace(r.URL.Query().Get("accepting_tasks"))
	if rawAccepting == "" {
		return query, true
	}

	value, err := strconv.ParseBool(rawAccepting)
	if err != nil {
		writeInvalidQuery(w, r, "accepting_tasks", "accepting_tasks must be a boolean")
		return workerapp.WorkerListQuery{}, false
	}
	query.AcceptingTasks = &value
	return query, true
}

func workerListResultFromApp(result workerapp.WorkerListResult) WorkerListResult {
	items := make([]WorkerView, 0, len(result.Items))
	for _, item := range result.Items {
		items = append(items, workerViewFromApp(item))
	}
	return WorkerListResult{
		Items:    items,
		Page:     result.Page,
		PageSize: result.PageSize,
		Total:    result.Total,
	}
}

func workerViewFromApp(view workerapp.WorkerView) WorkerView {
	tasks := make([]TaskView, 0, len(view.ActiveTasks))
	for _, task := range view.ActiveTasks {
		tasks = append(tasks, taskViewFromApp(task))
	}
	return WorkerView{
		WorkerID:        view.WorkerID,
		Namespace:       view.Namespace,
		PodName:         view.PodName,
		NodeName:        view.NodeName,
		ContainerName:   view.ContainerName,
		Image:           view.Image,
		Version:         view.Version,
		Status:          view.Status,
		StatusReason:    view.StatusReason,
		ProcessAlive:    view.ProcessAlive,
		AcceptingTasks:  view.AcceptingTasks,
		LastSeenAt:      view.LastSeenAt,
		LastReadyAt:     view.LastReadyAt,
		ActiveTaskCount: view.ActiveTaskCount,
		ActiveTasks:     tasks,
	}
}

func taskViewFromApp(view workerapp.TaskView) TaskView {
	return TaskView{
		TaskID:      view.TaskID,
		ExecPlanID:  view.ExecPlanID,
		TaskType:    view.TaskType,
		Phase:       view.Phase,
		PhaseName:   view.PhaseName,
		CurrentStep: stepViewFromApp(view.CurrentStep),
		StartedAt:   view.StartedAt,
		UpdatedAt:   view.UpdatedAt,
		Metadata:    cloneStringMap(view.Metadata),
	}
}

func taskDetailFromApp(view workerapp.TaskDetail) TaskDetail {
	return TaskDetail{
		TaskID:      view.TaskID,
		WorkerID:    view.WorkerID,
		ExecPlanID:  view.ExecPlanID,
		TaskType:    view.TaskType,
		Phase:       view.Phase,
		PhaseName:   view.PhaseName,
		CurrentStep: stepViewFromApp(view.CurrentStep),
		Status:      view.Status,
		StartedAt:   view.StartedAt,
		UpdatedAt:   view.UpdatedAt,
		EndedAt:     view.EndedAt,
		Metadata:    cloneStringMap(view.Metadata),
	}
}

func stepViewFromApp(step *workerapp.StepView) *StepView {
	if step == nil {
		return nil
	}
	return &StepView{
		Step:       step.Step,
		StepName:   step.StepName,
		Status:     step.Status,
		StartedAt:  step.StartedAt,
		UpdatedAt:  step.UpdatedAt,
		FinishedAt: step.FinishedAt,
		Attempt:    step.Attempt,
		ErrorClass: step.ErrorClass,
	}
}

func caseTimelineResultFromApp(result workerapp.CaseTimelineResult) CaseTimelineResult {
	items := make([]CaseStepView, 0, len(result.Items))
	for _, item := range result.Items {
		items = append(items, caseStepViewFromApp(item))
	}
	return CaseTimelineResult{
		TaskID: result.TaskID,
		Items:  items,
	}
}

func execPlanCaseDrilldownResultFromApp(result workerapp.ExecPlanCaseDrilldownResult) ExecPlanCaseDrilldownResult {
	items := make([]CaseStepView, 0, len(result.Items))
	for _, item := range result.Items {
		items = append(items, caseStepViewFromApp(item))
	}
	return ExecPlanCaseDrilldownResult{
		ExecPlanID: result.ExecPlanID,
		Items:      items,
		Page:       result.Page,
		PageSize:   result.PageSize,
		Total:      result.Total,
	}
}

func caseStepViewFromApp(view workerapp.CaseStepView) CaseStepView {
	return CaseStepView{
		TaskID:     view.TaskID,
		WorkerID:   view.WorkerID,
		ExecPlanID: view.ExecPlanID,
		Namespace:  view.Namespace,
		PodName:    view.PodName,
		NodeName:   view.NodeName,
		Step:       view.Step,
		StepName:   view.StepName,
		Status:     view.Status,
		Result:     view.Result,
		ErrorClass: view.ErrorClass,
		Attempt:    view.Attempt,
		StartedAt:  view.StartedAt,
		FinishedAt: view.FinishedAt,
		ObservedAt: view.ObservedAt,
		EventType:  view.EventType,
	}
}

func fleetSummaryFromApp(summary workerapp.FleetSummary) FleetSummary {
	return FleetSummary{
		TotalWorkers:     summary.TotalWorkers,
		OnlineWorkers:    summary.OnlineWorkers,
		DegradedWorkers:  summary.DegradedWorkers,
		OfflineWorkers:   summary.OfflineWorkers,
		UnknownWorkers:   summary.UnknownWorkers,
		AcceptingWorkers: summary.AcceptingWorkers,
		BusyWorkers:      summary.BusyWorkers,
		ActiveTaskCount:  summary.ActiveTaskCount,
	}
}

func alertListResultFromApp(result workerapp.AlertListResult) AlertListResult {
	items := make([]AlertView, 0, len(result.Items))
	for _, item := range result.Items {
		items = append(items, alertViewFromApp(item))
	}
	return AlertListResult{
		Items:    items,
		Page:     result.Page,
		PageSize: result.PageSize,
		Total:    result.Total,
	}
}

func alertViewFromApp(view workerapp.AlertView) AlertView {
	return AlertView{
		AlertID:     view.AlertID,
		WorkerID:    view.WorkerID,
		TaskID:      view.TaskID,
		AlertType:   view.AlertType,
		Status:      view.Status,
		Severity:    view.Severity,
		Message:     view.Message,
		TriggeredAt: view.TriggeredAt,
		ResolvedAt:  view.ResolvedAt,
	}
}

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

func parsePagination(w http.ResponseWriter, r *http.Request) (int, int, bool) {
	page := 1
	pageSize := 50

	if rawPage := strings.TrimSpace(r.URL.Query().Get("page")); rawPage != "" {
		value, err := strconv.Atoi(rawPage)
		if err != nil || value < 1 {
			writeInvalidQuery(w, r, "page", "page must be a positive integer")
			return 0, 0, false
		}
		page = value
	}

	if rawPageSize := strings.TrimSpace(r.URL.Query().Get("page_size")); rawPageSize != "" {
		value, err := strconv.Atoi(rawPageSize)
		if err != nil || value < 1 || value > 500 {
			writeInvalidQuery(w, r, "page_size", "page_size must be between 1 and 500")
			return 0, 0, false
		}
		pageSize = value
	}

	return page, pageSize, true
}

func writeInvalidQuery(w http.ResponseWriter, r *http.Request, field string, message string) {
	_ = contract.WriteError(w, r, contract.NewErrorBuilder().
		Type(contract.TypeValidation).
		Code(contract.CodeInvalidQuery).
		Message(message).
		Detail("field", field).
		Build())
}

func writeRequiredPathParam(w http.ResponseWriter, r *http.Request, field string) {
	_ = contract.WriteError(w, r, contract.NewErrorBuilder().
		Type(contract.TypeRequired).
		Message(field+" is required").
		Detail("field", field).
		Build())
}
