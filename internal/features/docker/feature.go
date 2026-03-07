package docker

import (
	"github.com/naodEthiop/lalibela-cli/internal/features/shared"
)

// Feature installs the "docker" scaffold feature.
type Feature struct{}

// New returns a new "docker" feature installer.
func New() Feature { return Feature{} }

// Name returns the registry name of the feature.
func (Feature) Name() string { return "docker" }

// Compatible reports whether the feature supports a given framework.
func (Feature) Compatible(framework string) bool {
	return shared.IsFeatureCompatible("docker", framework)
}

// Install writes the feature's scaffold files into projectRoot.
func (Feature) Install(projectRoot string) error {
	const file = `FROM golang:1.25-alpine AS builder
WORKDIR /src
COPY . .
RUN go build -o app .

FROM alpine:3.20
WORKDIR /app
COPY --from=builder /src/app /app/app
EXPOSE 8080
CMD ["/app/app"]
`
	return shared.WriteFileIfMissing(projectRoot, "deployments/Dockerfile", []byte(file))
}
