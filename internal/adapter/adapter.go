// Package adapter defines format capabilities and the boundary between source
// formats and canonical tables.
package adapter

import (
	"context"
	"fmt"
	"sync"

	"data-toolkit/internal/contract"
	"data-toolkit/internal/table"
)

type Capability string

const (
	CapabilityInspect         Capability = "inspect"
	CapabilityMapBasic        Capability = "map_basic"
	CapabilityMapExtended     Capability = "map_extended"
	CapabilityRead            Capability = "read"
	CapabilityWrite           Capability = "write"
	CapabilityValidate        Capability = "validate"
	CapabilitySheets          Capability = "sheets"
	CapabilityStyles          Capability = "styles"
	CapabilityDisplayedValues Capability = "displayed_values"
)

type Capabilities map[Capability]bool

func (capabilities Capabilities) Supports(required ...Capability) bool {
	for _, capability := range required {
		if !capabilities[capability] {
			return false
		}
	}
	return true
}

type SourceRequest struct {
	Source       contract.Source
	SnapshotPath string
	Mappings     []contract.FieldMapping
}

type Inspection struct {
	SourceID table.SourceID
	Sheets   []table.Sheet
}

type MappingMode string

const (
	MappingBasic    MappingMode = "basic"
	MappingExtended MappingMode = "extended"
)

type Finding struct {
	SourceID table.SourceID
	SheetID  table.SheetID
	ColumnID table.ColumnID
	Code     string
	Path     string
	Message  string
	Count    int64
}

type MappingReport struct {
	SourceID table.SourceID
	Tables   []table.Table
	Columns  []ColumnMapping
	Findings []Finding
}

type ColumnMapping struct {
	Column          table.Column
	ColumnID        table.ColumnID
	JSONPointer     string
	ValueCount      int64
	MissingCount    int64
	TypeCounts      map[table.ValueKind]int64
	ObservedFormats map[string]int64
	WhitespaceCount int64
}

type OutputRequest struct {
	Output     contract.Output
	StagedPath string
	Render     table.RenderOptions
}

type RowWriter interface {
	WriteRow(context.Context, table.Row) error
	Close() error
}

type Adapter interface {
	Format() contract.DataFormat
	Capabilities() Capabilities
	Inspect(context.Context, SourceRequest) (Inspection, error)
	Map(context.Context, SourceRequest, MappingMode) (MappingReport, error)
	OpenRows(context.Context, SourceRequest, table.Table) (table.RowStream, error)
	NewWriter(context.Context, OutputRequest, table.Table) (RowWriter, error)
	ValidateOutput(context.Context, OutputRequest) error
}

type Registry struct {
	mu       sync.RWMutex
	adapters map[contract.DataFormat]Adapter
}

func NewRegistry() *Registry {
	return &Registry{adapters: make(map[contract.DataFormat]Adapter)}
}

func (registry *Registry) Register(adapter Adapter) error {
	if adapter == nil {
		return fmt.Errorf("register adapter: adapter is nil")
	}
	format := adapter.Format()
	if format == "" {
		return fmt.Errorf("register adapter: format is required")
	}
	registry.mu.Lock()
	defer registry.mu.Unlock()
	if _, exists := registry.adapters[format]; exists {
		return fmt.Errorf("register adapter: format %q is already registered", format)
	}
	registry.adapters[format] = adapter
	return nil
}

func (registry *Registry) Require(format contract.DataFormat, required ...Capability) (Adapter, error) {
	registry.mu.RLock()
	adapter, exists := registry.adapters[format]
	registry.mu.RUnlock()
	if !exists {
		return nil, fmt.Errorf("adapter %q is not registered", format)
	}
	for _, capability := range required {
		if !adapter.Capabilities().Supports(capability) {
			return nil, fmt.Errorf("adapter %q does not support capability %q", format, capability)
		}
	}
	return adapter, nil
}
