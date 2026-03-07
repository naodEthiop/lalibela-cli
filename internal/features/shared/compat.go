package shared

import "strings"

// IsFeatureCompatible reports whether a feature should be offered/installed for
// a given framework.
func IsFeatureCompatible(featureName, framework string) bool {
	feature := strings.ToLower(strings.TrimSpace(featureName))
	fw := strings.ToLower(strings.TrimSpace(framework))

	switch fw {
	case "gin", "echo":
		return true
	case "fiber":
		// Swagger is intentionally excluded for Fiber auto-wiring.
		return feature != "swagger"
	case "nethttp":
		switch feature {
		case "docker", "logger", "postgres", "redis", "config", "graceful-shutdown", "health", "swagger":
			return true
		default:
			return false
		}
	default:
		return false
	}
}
