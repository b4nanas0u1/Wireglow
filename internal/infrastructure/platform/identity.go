package platform

// Identity отличает живой процесс от повторного использования его PID.
type Identity struct {
	PID   int    `json:"pid"`
	Start string `json:"start"`
	Path  string `json:"path"`
	Boot  string `json:"boot,omitempty"`
}
