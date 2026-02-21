package ratelimit

import (
	"github.com/naodEthiop/lalibela-cli/internal/features/shared"
)

type Feature struct{}

func New() Feature { return Feature{} }

func (Feature) Name() string { return "rate-limit" }

func (Feature) Compatible(framework string) bool {
	return shared.IsFeatureCompatible("rate-limit", framework)
}

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
