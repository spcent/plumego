package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/router"
	workerapp "workerfleet/internal/app"
	"workerfleet/internal/domain"
)

func (h *Handler) ListWorkers(w http.ResponseWriter, r *http.Request) {
	if !h.requireAdminAuth(w, r) {
		return
	}
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
	_ = contract.WriteResponse(w, r, http.StatusOK, result, nil)
}

func (h *Handler) GetWorker(w http.ResponseWriter, r *http.Request) {
	if !h.requireAdminAuth(w, r) {
		return
	}
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
	_ = contract.WriteResponse(w, r, http.StatusOK, result, nil)
}

func (h *Handler) GetTask(w http.ResponseWriter, r *http.Request) {
	if !h.requireAdminAuth(w, r) {
		return
	}
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
	_ = contract.WriteResponse(w, r, http.StatusOK, result, nil)
}

func (h *Handler) GetCaseTimeline(w http.ResponseWriter, r *http.Request) {
	if !h.requireAdminAuth(w, r) {
		return
	}
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
	_ = contract.WriteResponse(w, r, http.StatusOK, result, nil)
}

func (h *Handler) ListExecPlanCases(w http.ResponseWriter, r *http.Request) {
	if !h.requireAdminAuth(w, r) {
		return
	}
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
	_ = contract.WriteResponse(w, r, http.StatusOK, result, nil)
}

func (h *Handler) FleetSummary(w http.ResponseWriter, r *http.Request) {
	if !h.requireAdminAuth(w, r) {
		return
	}
	if h.service == nil {
		writeNotImplemented(w, r, "FLEET_SUMMARY_NOT_CONFIGURED", "fleet summary service not configured")
		return
	}

	result, err := h.service.FleetSummary(r.Context())
	if err != nil {
		writeServiceError(w, r, err)
		return
	}
	_ = contract.WriteResponse(w, r, http.StatusOK, result, nil)
}

func (h *Handler) ListAlerts(w http.ResponseWriter, r *http.Request) {
	if !h.requireAdminAuth(w, r) {
		return
	}
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
	_ = contract.WriteResponse(w, r, http.StatusOK, result, nil)
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
