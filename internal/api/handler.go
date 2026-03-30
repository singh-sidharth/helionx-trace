package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/singh-sidharth/helionx-trace/internal/model"
	"github.com/singh-sidharth/helionx-trace/internal/service"
	"github.com/singh-sidharth/helionx-trace/internal/store"
)

type Handler struct {
	store           store.EventStore
	timelineService *service.TimelineService
}

func NewHandler(store store.EventStore, timelineService *service.TimelineService) *Handler {
	return &Handler{
		store:           store,
		timelineService: timelineService,
	}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/health", h.handleHealth)
	mux.HandleFunc("/events", h.handlePostEvent)
	mux.HandleFunc("/debug/", h.handleDebug)
}

func (h *Handler) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
	})
}

func (h *Handler) handlePostEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	var event model.Event
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}

	if err := validateEvent(&event); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	if err := h.store.Add(event); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{
		"message": "event stored",
	})
}

func (h *Handler) handleDebug(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/debug/")
	path = strings.Trim(path, "/")

	if path == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "requestId is required"})
		return
	}

	if strings.HasSuffix(path, "/summary") {
		requestID := strings.TrimSuffix(path, "/summary")
		requestID = strings.Trim(requestID, "/")
		h.handleGetSummary(w, requestID)
		return
	}

	h.handleGetTimeline(w, path)
}

func (h *Handler) handleGetTimeline(w http.ResponseWriter, requestID string) {
	if requestID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "requestId is required"})
		return
	}

	resp, err := h.timelineService.Build(requestID)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) handleGetSummary(w http.ResponseWriter, requestID string) {
	if requestID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "requestId is required"})
		return
	}

	resp, err := h.timelineService.Build(requestID)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(formatSummary(resp)))
}

func formatSummary(resp model.TimelineResponse) string {
	var b strings.Builder

	fmt.Fprintf(&b, "Request ID: %s\n", resp.RequestID)
	fmt.Fprintf(&b, "Status: %s\n", resp.Status)

	if resp.FailurePoint != "" {
		fmt.Fprintf(&b, "Failure Point: %s\n", resp.FailurePoint)
	}

	fmt.Fprintf(&b, "Retry Count: %d\n", resp.RetryCount)
	fmt.Fprintf(&b, "Total Duration: %.3fs\n\n", float64(resp.TotalDurationMs)/1000.0)

	b.WriteString("Timeline\n")
	b.WriteString("--------\n")

	for i, step := range resp.Timeline {
		fmt.Fprintf(
			&b,
			"%d. [%s] %s -> %s",
			i+1,
			step.Service,
			step.EventType,
			step.Status,
		)

		if i > 0 {
			fmt.Fprintf(&b, " (+%.3fs)", float64(step.DeltaMs)/1000.0)
		}

		if errMsg := extractError(step); errMsg != "" {
			fmt.Fprintf(&b, " error=%s", errMsg)
		}

		if step.IsRetry {
			fmt.Fprintf(&b, " [retry]")
		}

		b.WriteString("\n")
	}

	return b.String()
}

func extractError(step model.TimelineStep) string {
	if step.Metadata == nil {
		return ""
	}

	raw, ok := step.Metadata["error"]
	if !ok {
		return ""
	}

	msg, ok := raw.(string)
	if !ok {
		return ""
	}

	return msg
}

func validateEvent(event *model.Event) error {
	if strings.TrimSpace(event.RequestID) == "" {
		return errors.New("requestId is required")
	}
	if strings.TrimSpace(event.Service) == "" {
		return errors.New("service is required")
	}
	if strings.TrimSpace(event.EventType) == "" {
		return errors.New("eventType is required")
	}
	if event.Status == "" {
		return errors.New("status is required")
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now().UTC()
	}
	return nil
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
