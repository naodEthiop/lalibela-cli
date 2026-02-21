package swagger

import (
	"github.com/naodEthiop/lalibela-cli/internal/features/shared"
)

type Feature struct{}

func New() Feature { return Feature{} }

func (Feature) Name() string { return "swagger" }

func (Feature) Compatible(framework string) bool {
	return shared.IsFeatureCompatible("swagger", framework)
}

func (Feature) Install(projectRoot string) error {
	const readme = `# Swagger Integration

This project was scaffolded with Lalibela swagger support.

## Notes

- Gin / Echo support direct middleware integration.
- Fiber is intentionally excluded from auto-wiring in Lalibela due common routing plugin conflicts.
- net/http requires manual route wiring for swagger UI and spec serving.

`
	return shared.WriteFileIfMissing(projectRoot, "docs/swagger/README.md", []byte(readme))
}
