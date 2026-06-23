package request_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"task-service/internal/adapter/http/pagination"
	"task-service/internal/adapter/http/request"
	"task-service/internal/adapter/http/requestctx"
	"task-service/internal/domain"
	"task-service/internal/domain/models"
)

func TestRegisterRequestValidateNormalizesInput(t *testing.T) {
	payload := request.RegisterRequest{
		Email:    " OWNER@EXAMPLE.COM ",
		Password: "password123",
		Name:     " Owner ",
	}

	if err := payload.Validate(); err != nil {
		t.Fatalf("validate failed: %v", err)
	}
	input := payload.ToInput()
	if input.Email != "owner@example.com" || input.Name != "Owner" {
		t.Fatalf("input was not normalized: %#v", input)
	}
}

func TestRegisterRequestValidateRejectsInvalidInput(t *testing.T) {
	cases := []request.RegisterRequest{
		{Email: "bad-email", Password: "password123", Name: "Owner"},
		{Email: "owner@example.com", Password: "short", Name: "Owner"},
		{Email: "owner@example.com", Password: "password123", Name: " "},
	}

	for _, tc := range cases {
		if err := tc.Validate(); !errors.Is(err, domain.ErrInvalidInput) {
			t.Fatalf("expected invalid input for %#v, got %v", tc, err)
		}
	}
}

func TestInviteRequestValidateDefaultsRole(t *testing.T) {
	payload := request.InviteRequest{Email: " MEMBER@EXAMPLE.COM "}

	if err := payload.Validate(); err != nil {
		t.Fatalf("validate failed: %v", err)
	}
	input := payload.ToInput()
	if input.Email != "member@example.com" || input.Role != models.RoleMember {
		t.Fatalf("input was not normalized: %#v", input)
	}
}

func TestInviteRequestValidateRejectsInvalidRole(t *testing.T) {
	cases := []request.InviteRequest{
		{Email: "member@example.com", Role: string(models.RoleOwner)},
		{Email: "member@example.com", Role: "bad"},
		{Email: "bad-email", Role: string(models.RoleMember)},
	}

	for _, tc := range cases {
		if err := tc.Validate(); !errors.Is(err, domain.ErrInvalidInput) {
			t.Fatalf("expected invalid input for %#v, got %v", tc, err)
		}
	}
}

func TestCreateTaskRequestValidateNormalizesAndDefaultsStatus(t *testing.T) {
	assigneeID := int64(2)
	payload := request.CreateTaskRequest{
		Title:       " Implement API ",
		Description: "  Add validation  ",
		TeamID:      1,
		AssigneeID:  &assigneeID,
	}

	if err := payload.Validate(); err != nil {
		t.Fatalf("validate failed: %v", err)
	}
	input := payload.ToInput()
	if input.Title != "Implement API" || input.Description != "Add validation" || input.Status != models.TaskStatusTodo {
		t.Fatalf("input was not normalized: %#v", input)
	}
}

func TestCreateTaskRequestValidateRejectsInvalidInput(t *testing.T) {
	badAssigneeID := int64(0)
	cases := []request.CreateTaskRequest{
		{Title: " ", TeamID: 1},
		{Title: "Task", TeamID: 0},
		{Title: "Task", TeamID: 1, Status: "bad"},
		{Title: "Task", TeamID: 1, AssigneeID: &badAssigneeID},
	}

	for _, tc := range cases {
		if err := tc.Validate(); !errors.Is(err, domain.ErrInvalidInput) {
			t.Fatalf("expected invalid input for %#v, got %v", tc, err)
		}
	}
}

func TestUpdateTaskRequestValidateRejectsInvalidInput(t *testing.T) {
	emptyTitle := " "
	badStatus := "bad"
	badAssigneeID := int64(-1)
	cases := []request.UpdateTaskRequest{
		{Title: &emptyTitle},
		{Status: &badStatus},
		{AssigneeID: request.OptionalJSON[int64]{Set: true, Valid: true, Value: badAssigneeID}},
	}

	for _, tc := range cases {
		if err := tc.Validate(); !errors.Is(err, domain.ErrInvalidInput) {
			t.Fatalf("expected invalid input for %#v, got %v", tc, err)
		}
	}
}

func TestUpdateTaskRequestToInputPreservesNullableFields(t *testing.T) {
	var payload request.UpdateTaskRequest
	if err := json.Unmarshal([]byte(`{"assignee_id":null,"due_date":null}`), &payload); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if err := payload.Validate(); err != nil {
		t.Fatalf("validate failed: %v", err)
	}

	input := payload.ToInput()
	if !input.AssigneeID.Set || input.AssigneeID.Valid {
		t.Fatalf("unexpected assignee optional state: %#v", input.AssigneeID)
	}
	if !input.DueDate.Set || input.DueDate.Valid {
		t.Fatalf("unexpected due_date optional state: %#v", input.DueDate)
	}

	var absent request.UpdateTaskRequest
	if err := json.Unmarshal([]byte(`{}`), &absent); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	input = absent.ToInput()
	if input.AssigneeID.Set || input.DueDate.Set {
		t.Fatalf("expected omitted fields to remain unset: %#v", input)
	}
}

func TestTaskFilterFromQueryParsesValidValues(t *testing.T) {
	createdAt := time.Date(2026, 6, 22, 10, 30, 0, 0, time.UTC)
	cursor := models.TaskCursor{CreatedAt: createdAt, ID: 100}
	req := httptest.NewRequest(http.MethodGet, "/tasks?team_id=1&status=todo&assignee_id=2&cursor="+pagination.EncodeTaskCursor(&cursor)+"&page_size=50", nil)

	filter, err := request.TaskFilterFromQuery(req)
	if err != nil {
		t.Fatalf("parse filter failed: %v", err)
	}
	if filter.TeamID == nil || *filter.TeamID != 1 {
		t.Fatalf("unexpected team filter: %#v", filter.TeamID)
	}
	if filter.Status == nil || *filter.Status != models.TaskStatusTodo {
		t.Fatalf("unexpected status filter: %#v", filter.Status)
	}
	if filter.AssigneeID == nil || *filter.AssigneeID != 2 {
		t.Fatalf("unexpected assignee filter: %#v", filter.AssigneeID)
	}
	if filter.Cursor == nil || !filter.Cursor.CreatedAt.Equal(createdAt) || filter.Cursor.ID != 100 {
		t.Fatalf("unexpected cursor: %#v", filter.Cursor)
	}
	if filter.PageSize != 50 {
		t.Fatalf("unexpected pagination: %#v", filter)
	}
}

func TestTaskFilterFromQueryRejectsInvalidValues(t *testing.T) {
	cases := []string{
		"/tasks?team_id=0",
		"/tasks?team_id=bad",
		"/tasks?status=bad",
		"/tasks?assignee_id=-1",
		"/tasks?page=1",
		"/tasks?cursor=bad",
		"/tasks?page_size=0",
		"/tasks?page_size=101",
	}

	for _, path := range cases {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		if _, err := request.TaskFilterFromQuery(req); !errors.Is(err, domain.ErrInvalidInput) {
			t.Fatalf("expected invalid input for %s, got %v", path, err)
		}
	}
}

func TestUserIDReadsAuthenticatedUser(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/tasks", nil)
	req = req.WithContext(requestctx.WithUserID(req.Context(), 42))

	userID, err := request.UserID(req)
	if err != nil {
		t.Fatalf("user id failed: %v", err)
	}
	if userID != 42 {
		t.Fatalf("unexpected user id: %d", userID)
	}
}

func TestUserIDRejectsMissingAuthenticatedUser(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/tasks", nil)

	if _, err := request.UserID(req); !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("expected unauthorized, got %v", err)
	}
}
