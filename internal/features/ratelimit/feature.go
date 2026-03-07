package ratelimit

import (
	"github.com/naodEthiop/lalibela-cli/internal/features/shared"
)

// Feature installs the "rate-limit" scaffold feature.
type Feature struct{}

// New returns a new "rate-limit" feature installer.
func New() Feature { return Feature{} }

// Name returns the registry name of the feature.
func (Feature) Name() string { return "rate-limit" }

// Compatible reports whether the feature supports a given framework.
func (Feature) Compatible(framework string) bool {
	return shared.IsFeatureCompatible("rate-limit", framework)
}

// Install writes the feature's scaffold files into projectRoot.
func (Feature) Install(projectRoot string) error {
	const file = `package server

import (
	"net/http"
	"time"

	"golang.org/x/time/rate"
)

func RateLimitMiddleware(rps int, burst int, next http.Handler) http.Handler {
	limiter := rate.NewLimiter(rate.Every(time.Second/time.Duration(rps)), burst)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !limiter.Allow() {
			http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}
`
	return shared.WriteFileIfMissing(projectRoot, "internal/server/rate_limit.go", []byte(file))
}
