package errorhandler

import (
	"github.com/naodEthiop/lalibela-cli/internal/features/shared"
)

type Feature struct{}

func New() Feature { return Feature{} }

func (Feature) Name() string { return "error-handler" }

func (Feature) Compatible(framework string) bool {
	return shared.IsFeatureCompatible("error-handler", framework)
}

func (Feature) Install(projectRoot string) error {
	const file = `package server

import (
	"encoding/json"
	"net/http"
)

type APIError struct {
	Code    string ` + "`json:\"code\"`" + `
	Message string ` + "`json:\"message\"`" + `
}

func WriteJSONError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(APIError{
		Code:    code,
		Message: message,
	})
}
`
	return shared.WriteFileIfMissing(projectRoot, "internal/server/error_handler.go", []byte(file))
}
