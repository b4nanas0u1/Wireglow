package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"wireglow/internal/infrastructure/tun"
)

func main() {
	uid := flag.Uint("uid", 0, "only permitted client UID")
	portable := flag.Bool("portable", false, "temporary elevated helper; no installed service")
	owner := flag.Int("owner", 0, "portable controller PID")
	flag.Parse()
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	if *portable {
		if err := tun.RunPortableHelper(ctx, *owner); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}
	server := tun.Server{UID: uint32(*uid), Binary: "/usr/local/libexec/wireglow-tun/mihomo", Runtime: "/run/wireglow-tun/private", Socket: tun.SocketPath}
	if err := server.Serve(ctx); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
