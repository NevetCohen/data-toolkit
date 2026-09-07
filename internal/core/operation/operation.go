// Package operation defines the shared V1 operation descriptor contract.
package operation

import (
	"encoding/json"

	"data-toolkit/internal/core"
)

// OperationDescriptor describes a registered Reader, Logical, or Writer
// operation for runtime capability discovery and workflow validation.
//
// The field order and JSON names are part of the V1 contract.
type OperationDescriptor struct {
	ID                   core.OperationID       `json:"id"`
	Version              string                 `json:"version"`
	Owner                core.OperationOwner    `json:"owner"`
	Exposure             core.OperationExposure `json:"exposure"`
	Name                 string                 `json:"name"`
	Description          string                 `json:"description"`
	Deterministic        bool                   `json:"deterministic"`
	StreamingMode        core.StreamingMode     `json:"streaming_mode"`
	InputShape           string                 `json:"input_shape"`
	OutputShape          string                 `json:"output_shape"`
	ParameterSchema      json.RawMessage        `json:"parameter_schema"`
	RequiredCapabilities []string               `json:"required_capabilities"`
}

// OperationRef identifies the resolved registered operation used by a plan.
type OperationRef struct {
	OperationID core.OperationID    `json:"operation_id"`
	Version     string              `json:"version"`
	Owner       core.OperationOwner `json:"owner"`
}
