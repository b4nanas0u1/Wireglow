package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"wireglow/internal/infrastructure/tun"
)

func main() {
	portable := flag.Bool("portable", false, "temporary elevated portable helper")
	owner := flag.Int("owner", 0, "controller PID")
	start := flag.String("owner-start", "", "controller creation time")
	flag.Parse()
	if !*portable || *owner <= 0 || *start == "" || flag.NArg() != 0 {
		fmt.Fprintln(os.Stderr, "tun: portable controller identity required")
		os.Exit(1)
	}
	if err := tun.RunWindowsPortableHelper(context.Background(), *owner, *start); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
