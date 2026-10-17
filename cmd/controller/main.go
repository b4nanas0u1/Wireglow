package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"

	"wireglow/internal/application/controller"
	"wireglow/internal/backend"
	"wireglow/internal/infrastructure/engines"
	"wireglow/internal/infrastructure/engines/manifest"
	"wireglow/internal/infrastructure/launcher"
	"wireglow/internal/infrastructure/network"
	"wireglow/internal/infrastructure/persistence"
	"wireglow/internal/infrastructure/platform"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	if len(os.Args) < 2 {
		return errors.New("usage: controller import|list|delete|connect|recover")
	}
	command := os.Args[1]
	fs := flag.NewFlagSet(command, flag.ContinueOnError)
	data := fs.String("data-dir", "", "private profile directory")
	runtime := fs.String("runtime-dir", "", "private runtime directory")
	id := fs.String("profile", "", "local profile ID")
	linkFile := fs.String("link-file", "", "file containing one FreeTurn link")
	relay := fs.String("relay", "", "absolute path to FreeTurn client")
	mihomo := fs.String("mihomo", "", "absolute path to mihomo "+manifest.Version("mihomo"))
	relayHash := fs.String("relay-sha256", "", "expected SHA256")
	mihomoHash := fs.String("mihomo-sha256", "", "expected SHA256")
	probe := fs.String("probe", "https://www.gstatic.com/generate_204", "HTTPS probe endpoint")
	probeStatus := fs.Int("probe-status", 204, "expected probe status")
	systemProxy := fs.Bool("system-proxy", false, "manage system proxy with recovery")
	runDir := fs.String("run-dir", "", "internal launcher directory")
	if err := fs.Parse(os.Args[2:]); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return errors.New("unexpected positional arguments")
	}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	ctx, stop := terminationContext(ctx)
	defer stop()
	if command == "worker" {
		return launcher.Worker(ctx, *runDir)
	}
	if command == "watchdog" {
		return launcher.Watchdog(ctx, *runDir)
	}
	paths, err := platform.DefaultPaths()
	if err != nil {
		return err
	}
	if *data != "" {
		paths.Data = *data
	}
	if *runtime != "" {
		paths.Runtime = *runtime
	}
	if err := platform.EnsurePaths(paths); err != nil {
		return err
	}
	if command == "desktop" {
		// EOF возникает также при аварии GUI; lease и recovery завершают сессию.
		go func() { _, _ = io.Copy(io.Discard, os.Stdin); cancel() }()
		return launcher.RunDesktop(ctx, backend.Options{Paths: paths, Relay: engines.Binary{Path: *relay, SHA256: *relayHash}, Mihomo: engines.Binary{Path: *mihomo, SHA256: *mihomoHash}, ProbeURL: *probe, ProbeStatus: *probeStatus}, func(endpoint launcher.Endpoint) { _ = json.NewEncoder(os.Stdout).Encode(endpoint) })
	}
	if command == "connect" {
		if *id == "" {
			return errors.New("--profile required")
		}
		return launcher.Run(ctx, backend.Options{Paths: paths, Relay: engines.Binary{Path: *relay, SHA256: *relayHash}, Mihomo: engines.Binary{Path: *mihomo, SHA256: *mihomoHash}, ProbeURL: *probe, ProbeStatus: *probeStatus, SystemProxy: *systemProxy}, *id, func(snapshot controller.Snapshot) { _ = json.NewEncoder(os.Stdout).Encode(snapshot) })
	}
	lock, err := platform.AcquireLock(filepath.Join(paths.Runtime, "controller.lock"))
	if err != nil {
		return err
	}
	defer lock.Close()
	store := &persistence.Profiles{Dir: filepath.Join(paths.Data, "profiles")}
	switch command {
	case "import":
		link, err := readLink(*linkFile)
		if err != nil {
			return err
		}
		p, err := store.Import(ctx, link)
		if err != nil {
			return err
		}
		return json.NewEncoder(os.Stdout).Encode(p)
	case "list":
		items, err := store.List(ctx)
		if err != nil {
			return err
		}
		return json.NewEncoder(os.Stdout).Encode(items)
	case "delete":
		return store.Delete(ctx, *id)
	case "recover":
		return (&network.Manager{Path: filepath.Join(paths.Data, "proxy-journal.json"), Settings: platform.ProxySettings{}}).Recover(ctx)
	default:
		return errors.New("unknown command")
	}
}

func readLink(path string) (string, error) {
	if path == "" {
		return "", errors.New("link file required")
	}
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	data, err := io.ReadAll(io.LimitReader(f, 512*1024+1))
	if err != nil {
		return "", err
	}
	if len(data) > 512*1024 {
		return "", errors.New("link file too large")
	}
	return string(data), nil
}
