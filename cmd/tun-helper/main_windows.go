package main

import (
	"fmt"
	"os"
)

func main() { fmt.Fprintln(os.Stderr, "TUN helper is not implemented on Windows yet"); os.Exit(1) }
