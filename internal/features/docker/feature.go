package docker

import (
	"github.com/naodEthiop/lalibela-cli/internal/features/shared"
)

type Feature struct{}

func New() Feature { return Feature{} }

func (Feature) Name() string { return "docker" }

func (Feature) Compatible(framework string) bool {
	return shared.IsFeatureCompatible("docker", framework)
}

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
