package manifest

import (
	"bytes"
	"strings"
	"testing"
)

func TestRejectInvalidManifest(t *testing.T) {
	for name, input := range map[string][]byte{
		"unknown field":           bytes.Replace(data, []byte("sha256:"), []byte("sha265:"), 1),
		"missing version":         bytes.Replace(data, []byte(`version: "`+Version("free-turn-proxy")+`"`), []byte(`version: ""`), 1),
		"invalid digest":          bytes.Replace(data, []byte("sha256: "), []byte("sha256: z"), 1),
		"unsupported compression": bytes.Replace(data, []byte("compression: none"), []byte("compression: zstd"), 1),
		"missing engine":          bytes.Replace(data, []byte("  mihomo:"), []byte("  other:"), 1),
		"multiple documents":      append(append([]byte(nil), data...), []byte("\n---\nengines: {}\n")...),
		"duplicate field":         bytes.Replace(data, []byte("    version:"), []byte("    version: duplicate\n    version:"), 1),
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := parse(input); err == nil {
				t.Fatal("invalid manifest accepted")
			}
		})
	}
}

func TestDownloaderFields(t *testing.T) {
	for name, engine := range pinned.Engines {
		for field, expected := range map[string]string{"version": Version(name), "repository": engine.Repository} {
			actual, err := Field(name + "." + field)
			if err != nil || actual != expected || actual == "" {
				t.Fatalf("%s.%s = %q, %v", name, field, actual, err)
			}
		}
		for platform, asset := range engine.Platforms {
			for field, expected := range map[string]string{"asset": asset.Name, "sha256": asset.SHA256, "compression": asset.Compression} {
				path := strings.Join([]string{name, "platforms", platform, field}, ".")
				actual, err := Field(path)
				if err != nil || actual != expected || actual == "" {
					t.Fatalf("%s = %q, %v", path, actual, err)
				}
			}
		}
	}
	for _, path := range []string{"unknown.version", "mihomo.unknown", "mihomo.platforms.unknown-amd64.asset", "mihomo.platforms.linux-amd64.sha265", "mihomo"} {
		if _, err := Field(path); err == nil {
			t.Errorf("unknown field %q accepted", path)
		}
	}
}
