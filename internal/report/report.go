// Package report defines deterministic application events and validation
// findings shared across the V1 foundation.
package report

type Event struct {
	Sequence  uint64 `json:"sequence"`
	Kind      string `json:"kind"`
	Component string `json:"component,omitempty"`
	Message   string `json:"message,omitempty"`
}

type Finding struct {
	Path    string `json:"path"`
	Code    string `json:"code"`
	Message string `json:"message"`
}
