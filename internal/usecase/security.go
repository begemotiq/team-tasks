package usecase

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"task-service/internal/domain"
	"task-service/internal/domain/models"
)

type PasswordHasher interface {
	Hash(password string) (string, error)
	Compare(hash string, password string) error
}

type BcryptHasher struct {
	Cost int
}

func (h BcryptHasher) Hash(password string) (string, error) {
	cost := h.Cost
	if cost == 0 {
		cost = bcrypt.DefaultCost
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func (h BcryptHasher) Compare(hash string, password string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		return domain.ErrUnauthorized
	}
	return nil
}

type TokenParser interface {
	ParseToken(tokenString string) (int64, error)
}

type JWTManager struct {
	Secret []byte
	TTL    time.Duration
}

type jwtClaims struct {
	UserID int64  `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

func (m JWTManager) NewToken(user *models.User) (string, error) {
	if len(m.Secret) == 0 {
		return "", fmt.Errorf("%w: jwt secret is empty", domain.ErrInvalidInput)
	}
	ttl := m.TTL
	if ttl == 0 {
		ttl = 24 * time.Hour
	}
	now := time.Now().UTC()
	claims := jwtClaims{
		UserID: user.ID,
		Email:  user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   fmt.Sprint(user.ID),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.Secret)
}

func (m JWTManager) ParseToken(tokenString string) (int64, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwtClaims{}, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, domain.ErrUnauthorized
		}
		return m.Secret, nil
	})
	if err != nil {
		return 0, domain.ErrUnauthorized
	}
	claims, ok := token.Claims.(*jwtClaims)
	if !ok || !token.Valid || claims.UserID == 0 {
		return 0, domain.ErrUnauthorized
	}
	return claims.UserID, nil
}
