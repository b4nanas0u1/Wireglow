// Package manifest exposes the embedded pins shared by the downloader and controller.
package manifest

import (
	"bytes"
	_ "embed"
	"encoding/hex"
	"fmt"
	"io"
	"regexp"
	"strings"

	"go.yaml.in/yaml/v3"
)

//go:embed manifest.yaml
var data []byte

// Engine describes an upstream release and its supported platform assets.
type Engine struct {
	Repository string           `yaml:"repository"`
	Version    string           `yaml:"version"`
	Platforms  map[string]Asset `yaml:"platforms"`
}

// Asset pins the downloaded file, before decompression.
type Asset struct {
	Name        string `yaml:"asset"`
	SHA256      string `yaml:"sha256"`
	Compression string `yaml:"compression"`
}

type document struct {
	Engines map[string]Engine `yaml:"engines"`
}

var pinned = func() document {
	m, err := parse(data)
	if err != nil {
		panic(fmt.Sprintf("engine manifest: %v", err))
	}
	return m
}()

func parse(input []byte) (document, error) {
	var m document
	decoder := yaml.NewDecoder(bytes.NewReader(input))
	decoder.KnownFields(true)
	if err := decoder.Decode(&m); err != nil {
		return m, err
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return m, fmt.Errorf("expected one YAML document")
	}
	for _, name := range []string{"free-turn-proxy", "mihomo"} {
		if _, ok := m.Engines[name]; !ok {
			return m, fmt.Errorf("missing engine %s", name)
		}
	}
	component := regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._+-]*$`)
	repository := regexp.MustCompile(`^[a-zA-Z0-9_.-]+/[a-zA-Z0-9_.-]+$`)
	for name, engine := range m.Engines {
		if !component.MatchString(name) || !repository.MatchString(engine.Repository) || !component.MatchString(engine.Version) || len(engine.Platforms) == 0 {
			return m, fmt.Errorf("invalid engine %s: repository, version and platforms required", name)
		}
		for platform, asset := range engine.Platforms {
			digest, err := hex.DecodeString(asset.SHA256)
			if !component.MatchString(platform) || !component.MatchString(asset.Name) || err != nil || len(digest) != 32 || (asset.Compression != "none" && asset.Compression != "gzip" && asset.Compression != "zip") {
				return m, fmt.Errorf("invalid asset %s/%s: name, SHA-256 and supported compression required", name, platform)
			}
		}
	}
	return m, nil
}

// Version returns the pinned release tag. Unknown names are programming errors.
func Version(name string) string {
	engine, ok := pinned.Engines[name]
	if !ok {
		panic("unknown engine: " + name)
	}
	return engine.Version
}

// Field returns a scalar for shell scripts without evaluating shell code.
// Paths are engine.version, engine.repository or engine.platforms.platform.field.
func Field(path string) (string, error) {
	// Asset names and versions can contain dots; only the lookup path is split.
	parts := strings.Split(path, ".")
	if len(parts) != 2 && len(parts) != 4 {
		return "", fmt.Errorf("invalid manifest field %q", path)
	}
	engine, ok := pinned.Engines[parts[0]]
	if ok {
		if len(parts) == 2 {
			switch parts[1] {
			case "version":
				return engine.Version, nil
			case "repository":
				return engine.Repository, nil
			}
		} else if parts[1] == "platforms" {
			if asset, ok := engine.Platforms[parts[2]]; ok {
				switch parts[3] {
				case "asset":
					return asset.Name, nil
				case "sha256":
					return asset.SHA256, nil
				case "compression":
					return asset.Compression, nil
				}
			}
		}
	}
	return "", fmt.Errorf("unknown manifest field %q", path)
}
