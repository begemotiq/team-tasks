package request

import (
	"strings"

	"task-service/internal/domain/models"
	teamcreateusecase "task-service/internal/usecase/team_create"
	teaminviteusecase "task-service/internal/usecase/team_invite"
)

type CreateTeamRequest struct {
	Name string `json:"name"`
}

func (r *CreateTeamRequest) Validate() error {
	name, err := requiredString("name", r.Name)
	if err != nil {
		return err
	}
	r.Name = name
	return nil
}

func (r CreateTeamRequest) ToInput() teamcreateusecase.Input {
	return teamcreateusecase.Input{Name: r.Name}
}

type InviteRequest struct {
	Email string `json:"email"`
	Role  string `json:"role"`
}

func (r *InviteRequest) Validate() error {
	email, err := normalizeEmail(r.Email)
	if err != nil {
		return err
	}
	r.Role = strings.TrimSpace(r.Role)
	if r.Role == "" {
		r.Role = string(models.RoleMember)
	}
	role := models.Role(r.Role)
	if role == models.RoleOwner || !role.Valid() {
		return invalidInput("invalid invite role")
	}
	r.Email = email
	return nil
}

func (r InviteRequest) ToInput() teaminviteusecase.Input {
	return teaminviteusecase.Input{
		Email: r.Email,
		Role:  models.Role(r.Role),
	}
}
