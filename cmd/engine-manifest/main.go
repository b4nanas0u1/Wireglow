// engine-manifest exposes embedded engine pins to build scripts.
package main

import (
	"fmt"
	"os"

	"wireglow/internal/infrastructure/engines/manifest"
)

func main() {
	if len(os.Args) != 3 || os.Args[1] != "get" {
		fmt.Fprintln(os.Stderr, "usage: engine-manifest get <engine.field>")
		os.Exit(2)
	}
	value, err := manifest.Field(os.Args[2])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println(value)
}
