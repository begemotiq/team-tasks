package request

import (
	authloginusecase "task-service/internal/usecase/auth_login"
	authregisterusecase "task-service/internal/usecase/auth_register"
)

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

func (r *RegisterRequest) Validate() error {
	email, err := normalizeEmail(r.Email)
	if err != nil {
		return err
	}
	name, err := requiredString("name", r.Name)
	if err != nil {
		return err
	}
	if err := requirePassword(r.Password, 8); err != nil {
		return err
	}
	r.Email = email
	r.Name = name
	return nil
}

func (r *RegisterRequest) ToInput() authregisterusecase.Input {
	return authregisterusecase.Input{
		Email:    r.Email,
		Password: r.Password,
		Name:     r.Name,
	}
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (r *LoginRequest) Validate() error {
	email, err := normalizeEmail(r.Email)
	if err != nil {
		return err
	}
	if err := requirePassword(r.Password, 1); err != nil {
		return err
	}
	r.Email = email
	return nil
}

func (r *LoginRequest) ToInput() authloginusecase.Input {
	return authloginusecase.Input{
		Email:    r.Email,
		Password: r.Password,
	}
}
