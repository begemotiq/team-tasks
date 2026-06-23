package request

import (
	"fmt"
	"net/mail"
	"strings"

	"task-service/internal/domain"
)

func invalidInput(format string, args ...any) error {
	return fmt.Errorf("%w: %s", domain.ErrInvalidInput, fmt.Sprintf(format, args...))
}

func requiredString(field string, value string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", invalidInput("%s is required", field)
	}
	return trimmed, nil
}

func normalizeEmail(value string) (string, error) {
	email := strings.ToLower(strings.TrimSpace(value))
	if email == "" {
		return "", invalidInput("email is required")
	}
	address, err := mail.ParseAddress(email)
	if err != nil || address.Address != email {
		return "", invalidInput("invalid email")
	}
	return email, nil
}

func requirePassword(value string, minLength int) error {
	if strings.TrimSpace(value) == "" {
		return invalidInput("password is required")
	}
	if len(value) < minLength {
		return invalidInput("password must be at least %d characters", minLength)
	}
	return nil
}
