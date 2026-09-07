// Package datatype defines immutable canonical values and the V1 data-type
// extension contract.
package datatype

import (
	"fmt"
	"sort"
	"strings"
	"sync"

	"data-toolkit/internal/core"
)

const CurrentVersion = "v1"

type ID = core.DataTypeID

// CanonicalKind identifies the canonical representation of a registered data type.
type CanonicalKind string

const (
	CanonicalKindNull    CanonicalKind = "null"
	CanonicalKindString  CanonicalKind = "string"
	CanonicalKindInteger CanonicalKind = "integer"
	CanonicalKindDecimal CanonicalKind = "decimal"
	CanonicalKindBoolean CanonicalKind = "boolean"
	CanonicalKindDate    CanonicalKind = "date"
	CanonicalKindTime    CanonicalKind = "time"
)

func (kind CanonicalKind) Validate() error {
	switch kind {
	case CanonicalKindNull, CanonicalKindString, CanonicalKindInteger,
		CanonicalKindDecimal, CanonicalKindBoolean, CanonicalKindDate, CanonicalKindTime:
		return nil
	default:
		return fmt.Errorf("canonical kind %q is invalid", kind)
	}
}

// DataTypeDescriptor describes one registered data type for runtime discovery.
type DataTypeDescriptor struct {
	ID            ID            `json:"id"`
	Version       string        `json:"version"`
	Name          string        `json:"name"`
	CanonicalKind CanonicalKind `json:"canonical_kind"`
}

// Descriptor is retained as a compatibility alias for the data-type extension contract.
type Descriptor = DataTypeDescriptor

type RenderOptions struct {
	NullDisplay      string
	DecimalPrecision int
	DateFormat       string
	TimeFormat       string
	PhoneFormat      string
}

// Handler is the complete compile-time extension template for one canonical
// data type.
type Handler interface {
	Descriptor() DataTypeDescriptor
	Parse(string) (core.Value, error)
	Validate(core.Value) error
	Compare(core.Value, core.Value) (int, error)
	Render(core.Value, RenderOptions) (string, error)
}

type Constructor func() (Handler, error)

type Registry struct {
	version  string
	mu       sync.RWMutex
	handlers map[ID]Handler
}

func NewRegistry(version string) *Registry {
	return &Registry{version: version, handlers: make(map[ID]Handler)}
}

func (registry *Registry) RegisterConstructor(constructor Constructor) error {
	if constructor == nil {
		return fmt.Errorf("register data type: constructor is nil")
	}
	handler, err := constructor()
	if err != nil {
		return fmt.Errorf("construct data type: %w", err)
	}
	return registry.Register(handler)
}

func (registry *Registry) Register(handler Handler) error {
	if registry == nil {
		return fmt.Errorf("register data type: registry is nil")
	}
	if handler == nil {
		return fmt.Errorf("register data type: handler is nil")
	}
	descriptor := handler.Descriptor()
	if err := validateDescriptor(descriptor, registry.version); err != nil {
		return fmt.Errorf("register data type: %w", err)
	}
	registry.mu.Lock()
	defer registry.mu.Unlock()
	key := normalizeID(descriptor.ID)
	if _, exists := registry.handlers[key]; exists {
		return fmt.Errorf("register data type: identity %q is already registered", descriptor.ID)
	}
	registry.handlers[key] = handler
	return nil
}

func (registry *Registry) Require(id ID) (Handler, error) {
	if registry == nil {
		return nil, fmt.Errorf("data-type registry is nil")
	}
	registry.mu.RLock()
	handler, exists := registry.handlers[normalizeID(id)]
	registry.mu.RUnlock()
	if !exists {
		return nil, fmt.Errorf("data type %q is not registered", id)
	}
	return handler, nil
}

func (registry *Registry) Descriptors() []DataTypeDescriptor {
	if registry == nil {
		return nil
	}
	registry.mu.RLock()
	result := make([]DataTypeDescriptor, 0, len(registry.handlers))
	for _, handler := range registry.handlers {
		result = append(result, handler.Descriptor())
	}
	registry.mu.RUnlock()
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result
}

func (registry *Registry) Parse(id ID, input string) (core.Value, error) {
	handler, err := registry.Require(id)
	if err != nil {
		return core.Value{}, err
	}
	value, err := handler.Parse(input)
	if err != nil {
		return core.Value{}, fmt.Errorf("parse %q value: %w", id, err)
	}
	if err := handler.Validate(value); err != nil {
		return core.Value{}, fmt.Errorf("validate parsed %q value: %w", id, err)
	}
	return value, nil
}

func (registry *Registry) Compare(left, right core.Value) (int, error) {
	if left.TypeID() != right.TypeID() {
		return 0, fmt.Errorf("cannot compare data types %q and %q", left.TypeID(), right.TypeID())
	}
	handler, err := registry.Require(left.TypeID())
	if err != nil {
		return 0, err
	}
	return handler.Compare(left, right)
}

func (registry *Registry) Render(value core.Value, options RenderOptions) (string, error) {
	handler, err := registry.Require(value.TypeID())
	if err != nil {
		return "", err
	}
	return handler.Render(value, options)
}

func validateDescriptor(descriptor Descriptor, version string) error {
	if err := descriptor.ID.Validate(); err != nil {
		return fmt.Errorf("identity: %w", err)
	}
	if descriptor.Version != version {
		return fmt.Errorf("version %q is incompatible with registry version %q", descriptor.Version, version)
	}
	if strings.TrimSpace(descriptor.Name) == "" {
		return fmt.Errorf("name is required for %q", descriptor.ID)
	}
	if err := descriptor.CanonicalKind.Validate(); err != nil {
		return err
	}
	return nil
}

func normalizeID(value ID) ID {
	return ID(strings.ToLower(strings.TrimSpace(string(value))))
}
