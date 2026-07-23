// Package format defines file-format descriptors and capability-specific V1
// extension interfaces.
package format

import (
	"context"

	"data-toolkit/internal/core"
	"data-toolkit/internal/core/table"
)

const CurrentVersion = "v1"

type Capability string

const (
	CapabilityInspect  Capability = "inspect"
	CapabilityMap      Capability = "map"
	CapabilityRead     Capability = "read"
	CapabilityWrite    Capability = "write"
	CapabilityStyle    Capability = "style"
	CapabilityValidate Capability = "validate"
)

type Descriptor struct {
	ID           string       `json:"id"`
	Version      string       `json:"version"`
	Name         string       `json:"name"`
	Extensions   []string     `json:"extensions,omitempty"`
	Capabilities []Capability `json:"capabilities,omitempty"`
}

type Provider interface {
	Descriptor() Descriptor
}

type Constructor func() (Provider, error)

type Source struct {
	ID       core.SourceID
	Location string
	Options  map[string]string
}

type InspectRequest struct {
	Source Source
}

type Inspection struct {
	SourceID core.SourceID
	Sheets   []string
}

type Inspector interface {
	Inspect(context.Context, InspectRequest) (Inspection, error)
}

type MapRequest struct {
	FormatID string
	Source   Source
	Extended bool
}

type Mapping struct {
	SourceID core.SourceID
	Tables   []table.Schema
	Findings []string
}

type Mapper interface {
	Map(context.Context, MapRequest) (Mapping, error)
}

type ReadRequest struct {
	Source Source
	Table  table.Schema
}

type Reader interface {
	OpenRows(context.Context, ReadRequest) (table.RowStream, error)
}

type WriteRequest struct {
	Location string
	Table    table.Schema
}

type RowWriter interface {
	WriteRow(context.Context, table.Row) error
	Close() error
}

type Writer interface {
	NewWriter(context.Context, WriteRequest) (RowWriter, error)
}

type StyleRequest struct {
	Location string
	Style    string
}

type Styler interface {
	ApplyStyle(context.Context, StyleRequest) error
}

type ValidateRequest struct {
	Location string
	Table    table.Schema
}

type Validator interface {
	ValidateOutput(context.Context, ValidateRequest) error
}
