// Package operation defines the V1 logical-operation extension template.
package operation

import (
	"context"
	"encoding/json"

	"data-toolkit/internal/core/table"
)

const CurrentVersion = "v1"

type Descriptor struct {
	Kind            string          `json:"kind"`
	Version         string          `json:"version"`
	Name            string          `json:"name"`
	Deterministic   bool            `json:"deterministic"`
	MinimumInputs   int             `json:"minimum_inputs"`
	MaximumInputs   int             `json:"maximum_inputs,omitempty"`
	ParameterSchema json.RawMessage `json:"parameter_schema"`
}

// Spec is the stable V1 workflow operation envelope. Operation-specific fields
// live only in Parameters.
type Spec struct {
	ID         string                     `json:"id" yaml:"id"`
	Kind       string                     `json:"kind" yaml:"kind"`
	Inputs     []string                   `json:"inputs" yaml:"inputs"`
	Parameters json.RawMessage            `json:"parameters" yaml:"parameters"`
	Overrides  map[string]json.RawMessage `json:"overrides,omitempty" yaml:"overrides,omitempty"`
}

type Request struct {
	Spec   Spec
	Inputs []table.Dataset
}

type Result struct {
	Dataset table.Dataset
}

// Executor validates metadata and parameters before consuming rows, then
// executes deterministically over canonical datasets.
type Executor interface {
	Descriptor() Descriptor
	Validate(Spec, []table.Schema) error
	Execute(context.Context, Request) (Result, error)
}

type Constructor func() (Executor, error)
