package request

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"task-service/internal/adapter/http/pagination"
	"task-service/internal/adapter/http/requestctx"
	"task-service/internal/domain"
	"task-service/internal/domain/models"
)

func TaskFilterFromQuery(r *http.Request) (models.TaskFilter, error) {
	values := r.URL.Query()
	var filter models.TaskFilter
	if value := values.Get("team_id"); value != "" {
		parsed, err := strconv.ParseInt(value, 10, 64)
		if err != nil || parsed <= 0 {
			return models.TaskFilter{}, invalidInput("team_id must be a positive integer")
		}
		filter.TeamID = &parsed
	}
	if value := values.Get("status"); value != "" {
		status := models.TaskStatus(value)
		if !status.Valid() {
			return models.TaskFilter{}, invalidInput("invalid task status")
		}
		filter.Status = &status
	}
	if value := values.Get("assignee_id"); value != "" {
		parsed, err := strconv.ParseInt(value, 10, 64)
		if err != nil || parsed <= 0 {
			return models.TaskFilter{}, invalidInput("assignee_id must be a positive integer")
		}
		filter.AssigneeID = &parsed
	}
	if values.Get("page") != "" {
		return models.TaskFilter{}, invalidInput("page is not supported, use cursor")
	}
	cursor, err := pagination.DecodeTaskCursor(values.Get("cursor"))
	if err != nil {
		return models.TaskFilter{}, err
	}
	pageSize, err := intFromQuery(values.Get("page_size"), "page_size", 20, 100)
	if err != nil {
		return models.TaskFilter{}, err
	}
	filter.Cursor = cursor
	filter.PageSize = pageSize
	return filter.Normalize(), nil
}

func PathID(r *http.Request, name string) (int64, error) {
	value := chi.URLParam(r, name)
	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil || id <= 0 {
		return 0, invalidInput("%s must be a positive integer", name)
	}
	return id, nil
}

func UserID(r *http.Request) (int64, error) {
	userID, ok := requestctx.UserIDFromContext(r.Context())
	if !ok {
		return 0, domain.ErrUnauthorized
	}
	return userID, nil
}

func intFromQuery(value string, field string, fallback int, max int) (int, error) {
	if value == "" {
		return fallback, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 1 {
		return 0, invalidInput("%s must be a positive integer", field)
	}
	if max > 0 && parsed > max {
		return 0, invalidInput("%s must be less than or equal to %d", field, max)
	}
	return parsed, nil
}
