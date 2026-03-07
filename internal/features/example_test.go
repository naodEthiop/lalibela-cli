package features_test

import (
	"fmt"
	"os"

	"github.com/naodEthiop/lalibela-cli/internal/features"
)

func ExampleInstallFeature() {
	tmp, err := os.MkdirTemp("", "lalibela-features-example-*")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer os.RemoveAll(tmp)

	result, err := features.InstallFeature(tmp, "gin", "docker", nil)
	fmt.Println(result.Installed, result.AlreadyPresent, result.Compatible, err == nil)
	// Output: true false true true
}
