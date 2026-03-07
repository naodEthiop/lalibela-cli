package lalibelacli_test

import (
	"fmt"
	"io/fs"

	lalibelacli "github.com/naodEthiop/lalibela-cli"
)

func ExampleEmbeddedTemplates() {
	body, err := fs.ReadFile(lalibelacli.EmbeddedTemplates, "index.html")
	fmt.Println(err == nil && len(body) > 0)
	// Output: true
}
