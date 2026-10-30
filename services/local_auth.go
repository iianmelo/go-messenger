package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.messenger/config"
	"go.messenger/models"
	"golang.org/x/crypto/bcrypt"
)

type sessionClaims struct {
	UID string `json:"uid"`
	jwt.RegisteredClaims
}

func HashPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashed), nil
}

func CheckPassword(hash, password string) bool {
	if hash == "" || password == "" {
		return false
	}
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

func IssueToken(user models.User) (string, error) {
	if user.FireToken == "" {
		return "", errors.New("user has no session id")
	}
	claims := sessionClaims{
		UID: user.FireToken,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   fmt.Sprint(user.ID),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(72 * time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.JWTSecret()))
}

func ParseToken(raw string) (string, error) {
	parsed, err := jwt.ParseWithClaims(raw, &sessionClaims{}, func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(config.JWTSecret()), nil
	})
	if err != nil {
		return "", err
	}
	claims, ok := parsed.Claims.(*sessionClaims)
	if !ok || !parsed.Valid || claims.UID == "" {
		return "", errors.New("invalid token")
	}
	return claims.UID, nil
}

func PublicUser(user models.User) map[string]interface{} {
	return map[string]interface{}{
		"id":     user.ID,
		"name":   user.Name,
		"email":  user.Email,
		"avatar": user.Avatar,
	}
}
