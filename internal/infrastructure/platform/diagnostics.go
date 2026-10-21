package platform

import (
	"strings"
	"sync"
)

// EngineDiagnostics discards engine output after classifying fixed error
// phrases. Raw lines can contain credentials and must never be published.
type EngineDiagnostics struct {
	mu       sync.Mutex
	line     []byte
	dropping bool
	labels   map[string]bool
}

func (d *EngineDiagnostics) Write(data []byte) (int, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	for _, b := range data {
		if b == '\n' {
			if !d.dropping {
				d.classify(string(d.line))
			}
			d.line = nil
			d.dropping = false
		} else if !d.dropping {
			if len(d.line) >= 4096 {
				d.line = nil
				d.dropping = true
			} else {
				d.line = append(d.line, b)
			}
		}
	}
	return len(data), nil
}
func (d *EngineDiagnostics) classify(line string) {
	line = strings.ToLower(line)
	if !strings.Contains(line, "error") && !strings.Contains(line, "fail") && !strings.Contains(line, "timeout") && !strings.Contains(line, "eof") {
		return
	}
	label := ""
	switch {
	case strings.Contains(line, "resolve") || strings.Contains(line, "dns"):
		label = "DNS resolution failed"
	case strings.Contains(line, "certificate"):
		label = "certificate verification failed"
	case strings.Contains(line, "connection refused"):
		label = "outbound connection refused"
	case strings.Contains(line, "network is unreachable") || strings.Contains(line, "host is unreachable"):
		label = "outbound network unreachable"
	case strings.Contains(line, "timeout") || strings.Contains(line, "timed out") || strings.Contains(line, "deadline exceeded"):
		label = "outbound timed out"
	case strings.Contains(line, "connection reset"):
		label = "outbound connection reset"
	case strings.Contains(line, "eof"):
		label = "outbound peer closed connection"
	case strings.Contains(line, "dial"):
		label = "outbound dial failed"
	case strings.Contains(line, "handshake"):
		label = "outbound handshake failed"
	}
	if label != "" {
		if d.labels == nil {
			d.labels = make(map[string]bool)
		}
		d.labels[label] = true
	}
}
func (d *EngineDiagnostics) Summary() string {
	d.mu.Lock()
	defer d.mu.Unlock()
	var result []string
	for _, label := range []string{"DNS resolution failed", "certificate verification failed", "outbound connection refused", "outbound network unreachable", "outbound timed out", "outbound connection reset", "outbound peer closed connection", "outbound dial failed", "outbound handshake failed"} {
		if d.labels[label] {
			result = append(result, label)
		}
	}
	return strings.Join(result, ", ")
}

func (d *EngineDiagnostics) Reset() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.labels = nil
	// Discard a partial line; it may belong to the previous startup stage.
	if len(d.line) > 0 {
		d.line = nil
		d.dropping = true
	}
}
