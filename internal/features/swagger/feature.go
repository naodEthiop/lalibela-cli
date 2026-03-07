package swagger

import (
	"github.com/naodEthiop/lalibela-cli/internal/features/shared"
)

// Feature installs the "swagger" scaffold feature.
type Feature struct{}

// New returns a new "swagger" feature installer.
func New() Feature { return Feature{} }

// Name returns the registry name of the feature.
func (Feature) Name() string { return "swagger" }

// Compatible reports whether the feature supports a given framework.
func (Feature) Compatible(framework string) bool {
	return shared.IsFeatureCompatible("swagger", framework)
}

// Install writes the feature's scaffold files into projectRoot.
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
