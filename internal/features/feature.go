package features

// Feature describes an optional scaffold feature that can be installed into a
// generated project (for example: logger, postgres, docker).
type Feature interface {
	Name() string
	Compatible(framework string) bool
	Install(projectRoot string) error
}

// CommandRunner runs an external command in the given directory.
type CommandRunner = func(dir string, name string, args ...string) error

// InstallResult describes the outcome of attempting to install a feature.
type InstallResult struct {
	Name           string
	Installed      bool
	AlreadyPresent bool
	Compatible     bool
}
