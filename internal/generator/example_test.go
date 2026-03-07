package generator_test

import (
	"fmt"

	"github.com/naodEthiop/lalibela-cli/internal/generator"
)

func ExampleNormalizeFeatureNames() {
	features, err := generator.NormalizeFeatureNames([]string{"logger", "jwt", "docker"})
	fmt.Println(features)
	fmt.Println(err)
	// Output:
	// [Logger JWT Docker]
	// <nil>
}
