package ui

// FrameworkIcon returns a short icon for a framework identifier.
func FrameworkIcon(framework string) string {
	switch framework {
	case "gin":
		return "🥃"
	case "echo":
		return "📡"
	case "fiber":
		return "⚡"
	case "nethttp":
		return "🌐"
	default:
		return "•"
	}
}

// FrameworkLabel returns a human-friendly label for a framework identifier.
func FrameworkLabel(framework string) string {
	if framework == "nethttp" {
		return "net/http"
	}
	return framework
}

// FrameworkDescription returns a short description for a framework identifier.
func FrameworkDescription(framework string) string {
	switch framework {
	case "gin":
		return "Fast HTTP web framework"
	case "echo":
		return "High performance minimalist"
	case "fiber":
		return "Express inspired framework"
	case "nethttp":
		return "Standard library HTTP"
	default:
		return "Framework option"
	}
}
