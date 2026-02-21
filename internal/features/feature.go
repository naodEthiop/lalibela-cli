package features

type Feature interface {
	Name() string
	Compatible(framework string) bool
	Install(projectRoot string) error
}

type CommandRunner = func(dir string, name string, args ...string) error

type InstallResult struct {
	Name           string
	Installed      bool
	AlreadyPresent bool
	Compatible     bool
}
