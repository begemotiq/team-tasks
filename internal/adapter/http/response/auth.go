package response

import "task-service/internal/domain/models"

type AuthResponse struct {
	User  UserResponse `json:"user"`
	Token string       `json:"token"`
}

func NewAuth(user models.User, token string) AuthResponse {
	return AuthResponse{
		User:  NewUser(user),
		Token: token,
	}
}
