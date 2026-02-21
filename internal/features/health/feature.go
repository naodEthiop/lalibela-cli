package health

import (
	"github.com/naodEthiop/lalibela-cli/internal/features/shared"
)

type Feature struct{}

func New() Feature { return Feature{} }

func (Feature) Name() string { return "health" }

func (Feature) Compatible(framework string) bool {
	return shared.IsFeatureCompatible("health", framework)
}

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
