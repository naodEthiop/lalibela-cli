package auth

import (
	"github.com/naodEthiop/lalibela-cli/internal/features/shared"
)

type Feature struct{}

func New() Feature { return Feature{} }

func (Feature) Name() string { return "auth" }

func (Feature) Compatible(framework string) bool {
	return shared.IsFeatureCompatible("auth", framework)
}

func (Feature) Install(projectRoot string) error {
	const file = `package server

import (
	"errors"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

func ValidateJWT(tokenString string, secret []byte) error {
	if strings.TrimSpace(tokenString) == "" {
		return errors.New("missing bearer token")
	}
	_, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		return secret, nil
	})
	return err
}
`
	return shared.WriteFileIfMissing(projectRoot, "internal/server/auth_jwt.go", []byte(file))
}
