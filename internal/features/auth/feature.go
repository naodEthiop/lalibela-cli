package auth

import (
	"github.com/naodEthiop/lalibela-cli/internal/features/shared"
)

// Feature installs the "auth" scaffold feature.
type Feature struct{}

// New returns a new "auth" feature installer.
func New() Feature { return Feature{} }

// Name returns the registry name of the feature.
func (Feature) Name() string { return "auth" }

// Compatible reports whether the feature supports a given framework.
func (Feature) Compatible(framework string) bool {
	return shared.IsFeatureCompatible("auth", framework)
}

// Install writes the feature's scaffold files into projectRoot.
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
