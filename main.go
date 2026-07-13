// main.go

package main

import (
	"fmt"
	"os"

	"github.com/bgreenwell/git-ego/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}
}
