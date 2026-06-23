package http

import (
	"log/slog"
	stdhttp "net/http"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	authlogin "task-service/internal/adapter/http/handlers/auth_login"
	authregister "task-service/internal/adapter/http/handlers/auth_register"
	"task-service/internal/adapter/http/handlers/health"
	reportinvalidassignees "task-service/internal/adapter/http/handlers/report_invalid_assignees"
	reportteamsummary "task-service/internal/adapter/http/handlers/report_team_summary"
	reporttopcreators "task-service/internal/adapter/http/handlers/report_top_creators"
	taskcreate "task-service/internal/adapter/http/handlers/task_create"
	taskhistory "task-service/internal/adapter/http/handlers/task_history"
	tasklist "task-service/internal/adapter/http/handlers/task_list"
	taskupdate "task-service/internal/adapter/http/handlers/task_update"
	teamcreate "task-service/internal/adapter/http/handlers/team_create"
	teamdelete "task-service/internal/adapter/http/handlers/team_delete"
	teaminvite "task-service/internal/adapter/http/handlers/team_invite"
	teamlist "task-service/internal/adapter/http/handlers/team_list"
	"task-service/internal/usecase"
	authloginusecase "task-service/internal/usecase/auth_login"
	authregisterusecase "task-service/internal/usecase/auth_register"
	reportinvalidassigneesusecase "task-service/internal/usecase/report_invalid_assignees"
	reportteamsummaryusecase "task-service/internal/usecase/report_team_summary"
	reporttopcreatorsusecase "task-service/internal/usecase/report_top_creators"
	taskcreateusecase "task-service/internal/usecase/task_create"
	taskhistoryusecase "task-service/internal/usecase/task_history"
	tasklistusecase "task-service/internal/usecase/task_list"
	taskupdateusecase "task-service/internal/usecase/task_update"
	teamcreateusecase "task-service/internal/usecase/team_create"
	teamdeleteusecase "task-service/internal/usecase/team_delete"
	teaminviteusecase "task-service/internal/usecase/team_invite"
	teamlistusecase "task-service/internal/usecase/team_list"
)

type Dependencies struct {
	AuthRegister           *authregisterusecase.UseCase
	AuthLogin              *authloginusecase.UseCase
	TeamCreate             *teamcreateusecase.UseCase
	TeamDelete             *teamdeleteusecase.UseCase
	TeamList               *teamlistusecase.UseCase
	TeamInvite             *teaminviteusecase.UseCase
	TaskCreate             *taskcreateusecase.UseCase
	TaskList               *tasklistusecase.UseCase
	TaskUpdate             *taskupdateusecase.UseCase
	TaskHistory            *taskhistoryusecase.UseCase
	ReportTeamSummary      *reportteamsummaryusecase.UseCase
	ReportTopCreators      *reporttopcreatorsusecase.UseCase
	ReportInvalidAssignees *reportinvalidassigneesusecase.UseCase
	Tokens                 usecase.TokenParser
	RequestsPerMinute      int
	Logger                 *slog.Logger
}

func NewRouter(deps Dependencies) stdhttp.Handler {
	logger := deps.Logger
	if logger == nil {
		logger = discardLogger()
	}
	limiter := NewRateLimiter(deps.RequestsPerMinute)
	healthHandler := health.New()
	registerHandler := authregister.New(deps.AuthRegister)
	loginHandler := authlogin.New(deps.AuthLogin)
	createTeamHandler := teamcreate.New(deps.TeamCreate)
	deleteTeamHandler := teamdelete.New(deps.TeamDelete)
	listTeamsHandler := teamlist.New(deps.TeamList)
	inviteHandler := teaminvite.New(deps.TeamInvite)
	createTaskHandler := taskcreate.New(deps.TaskCreate)
	listTasksHandler := tasklist.New(deps.TaskList)
	updateTaskHandler := taskupdate.New(deps.TaskUpdate)
	taskHistoryHandler := taskhistory.New(deps.TaskHistory)
	teamSummaryHandler := reportteamsummary.New(deps.ReportTeamSummary)
	topCreatorsHandler := reporttopcreators.New(deps.ReportTopCreators)
	invalidAssigneesHandler := reportinvalidassignees.New(deps.ReportInvalidAssignees)

	r := chi.NewRouter()
	r.Use(requestIDMiddleware)
	r.Use(requestLogMiddleware(logger))
	r.Use(metricsMiddleware)
	r.Use(recoverMiddleware(logger))

	r.Handle("/metrics", promhttp.Handler())
	r.Get("/healthz", healthHandler.Handle)
	r.Get("/swagger", func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		stdhttp.Redirect(w, r, "/swagger/", stdhttp.StatusMovedPermanently)
	})
	r.Get("/swagger/", swaggerUI)
	r.Get("/swagger/openapi.yaml", openAPIYAML)

	r.Route("/api/v1", func(r chi.Router) {
		r.With(rateLimitMiddleware(limiter)).Post("/register", registerHandler.Handle)
		r.With(rateLimitMiddleware(limiter)).Post("/login", loginHandler.Handle)

		r.Group(func(r chi.Router) {
			r.Use(rateLimitMiddleware(limiter))
			r.Use(authMiddleware(deps.Tokens))

			r.Post("/teams", createTeamHandler.Handle)
			r.Get("/teams", listTeamsHandler.Handle)
			r.Delete("/teams/{id}", deleteTeamHandler.Handle)
			r.Post("/teams/{id}/invite", inviteHandler.Handle)

			r.Post("/tasks", createTaskHandler.Handle)
			r.Get("/tasks", listTasksHandler.Handle)
			r.Put("/tasks/{id}", updateTaskHandler.Handle)
			r.Get("/tasks/{id}/history", taskHistoryHandler.Handle)

			r.Get("/reports/team-summary", teamSummaryHandler.Handle)
			r.Get("/reports/top-creators", topCreatorsHandler.Handle)
			r.Get("/reports/invalid-assignees", invalidAssigneesHandler.Handle)
		})
	})

	return r
}
