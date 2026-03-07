package errorhandler

import (
	"github.com/naodEthiop/lalibela-cli/internal/features/shared"
)

// Feature installs the "error-handler" scaffold feature.
type Feature struct{}

// New returns a new "error-handler" feature installer.
func New() Feature { return Feature{} }

// Name returns the registry name of the feature.
func (Feature) Name() string { return "error-handler" }

// Compatible reports whether the feature supports a given framework.
func (Feature) Compatible(framework string) bool {
	return shared.IsFeatureCompatible("error-handler", framework)
}

// Install writes the feature's scaffold files into projectRoot.
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
