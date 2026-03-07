package health

import (
	"github.com/naodEthiop/lalibela-cli/internal/features/shared"
)

// Feature installs the "health" scaffold feature.
type Feature struct{}

// New returns a new "health" feature installer.
func New() Feature { return Feature{} }

// Name returns the registry name of the feature.
func (Feature) Name() string { return "health" }

// Compatible reports whether the feature supports a given framework.
func (Feature) Compatible(framework string) bool {
	return shared.IsFeatureCompatible("health", framework)
}

// Install writes the feature's scaffold files into projectRoot.
func (Feature) Install(projectRoot string) error {
	const file = `package server

import (
	"encoding/json"
	"net/http"
	"time"
)

type HealthResponse struct {
	Status    string ` + "`json:\"status\"`" + `
	Timestamp string ` + "`json:\"timestamp\"`" + `
}

func WriteHealth(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(HealthResponse{
		Status:    "ok",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	})
}
`
	return shared.WriteFileIfMissing(projectRoot, "internal/server/health.go", []byte(file))
}
