package platform

import (
	"os"
	"strings"
)

func withoutProxyEnvironment() []string {
	var result []string
	for _, entry := range os.Environ() {
		key, _, _ := strings.Cut(entry, "=")
		switch strings.ToLower(key) {
		case "http_proxy", "https_proxy", "all_proxy", "no_proxy":
			continue
		}
		result = append(result, entry)
	}
	return result
}
