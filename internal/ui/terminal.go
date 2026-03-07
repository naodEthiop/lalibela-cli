package ui

// InitTerminal initializes terminal features (such as ANSI escape support on
// Windows) used by the CLI UI.
func InitTerminal() {
	_ = enableVirtualTerminal()
}
