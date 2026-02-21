package cors

import (
	"github.com/naodEthiop/lalibela-cli/internal/features/shared"
)

type Feature struct{}

func New() Feature { return Feature{} }

func (Feature) Name() string { return "cors" }

func (Feature) Compatible(framework string) bool {
	return shared.IsFeatureCompatible("cors", framework)
}

func (Feature) Install(projectRoot string) error {
	const file = `package server

import "net/http"

func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type,Authorization")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
`
	return shared.WriteFileIfMissing(projectRoot, "internal/server/cors.go", []byte(file))
}
