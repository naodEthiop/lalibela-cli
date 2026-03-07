package cli_test

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/naodEthiop/lalibela-cli/internal/cli"
)

func ExampleParseArgs() {
	tmp, err := os.MkdirTemp("", "lalibela-cli-example-*")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer os.RemoveAll(tmp)

	// Use a non-existent config path so the output is deterministic.
	configPath := filepath.Join(tmp, "missing.json")

	opts, err := cli.ParseArgs([]string{
		"-config", configPath,
		"-name", "myapp",
		"-framework", "gin",
		"-features", "logger,jwt",
		"-yes",
	})
	fmt.Println(err == nil, opts.ProjectName, opts.Framework, opts.Features, opts.AssumeYes)
	// Output: true myapp gin [Logger JWT] true
}
