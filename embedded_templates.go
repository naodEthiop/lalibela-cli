package lalibelacli

import "embed"

// EmbeddedTemplates ships scaffold templates inside the binary so go-installed
// builds can generate projects without an external templates directory.
//
//go:embed templates
var EmbeddedTemplates embed.FS
