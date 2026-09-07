// Package style defines the descriptor exposed for registered table styles.
package style

// StyleDescriptor describes one named, configured table style.
type StyleDescriptor struct {
	Name        string `json:"name"`
	RightToLeft bool   `json:"right_to_left"`
}
