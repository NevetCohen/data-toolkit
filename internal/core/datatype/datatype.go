// Package datatype defines immutable canonical values and the V1 data-type
// extension contract.
package datatype

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
	"sync"
)

const CurrentVersion = "v1"

type ID string

type Descriptor struct {
	ID      ID     `json:"id"`
	Version string `json:"version"`
	Name    string `json:"name"`
}

type RenderOptions struct {
	NullDisplay      string
	DecimalPrecision int
	DateFormat       string
	TimeFormat       string
	PhoneFormat      string
}

// Value is immutable. Constructor and accessor methods copy encoded bytes.
type Value struct {
	typeID  ID
	encoded []byte
}

func NewValue(typeID ID, encoded []byte) (Value, error) {
	if strings.TrimSpace(string(typeID)) == "" {
		return Value{}, fmt.Errorf("data type identity is required")
	}
	if len(encoded) == 0 {
		return Value{}, fmt.Errorf("canonical encoded value is required")
	}
	return Value{typeID: typeID, encoded: append([]byte(nil), encoded...)}, nil
}

func (value Value) TypeID() ID {
	return value.typeID
}

func (value Value) Encoded() []byte {
	return append([]byte(nil), value.encoded...)
}

func (value Value) Equal(other Value) bool {
	return value.typeID == other.typeID && bytes.Equal(value.encoded, other.encoded)
}

// Handler is the complete compile-time extension template for one canonical
// data type.
type Handler interface {
	Descriptor() Descriptor
	Parse(string) (Value, error)
	Validate(Value) error
	Compare(Value, Value) (int, error)
	Render(Value, RenderOptions) (string, error)
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

func (registry *Registry) Descriptors() []Descriptor {
	if registry == nil {
		return nil
	}
	registry.mu.RLock()
	result := make([]Descriptor, 0, len(registry.handlers))
	for _, handler := range registry.handlers {
		result = append(result, handler.Descriptor())
	}
	registry.mu.RUnlock()
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result
}

func (registry *Registry) Parse(id ID, input string) (Value, error) {
	handler, err := registry.Require(id)
	if err != nil {
		return Value{}, err
	}
	value, err := handler.Parse(input)
	if err != nil {
		return Value{}, fmt.Errorf("parse %q value: %w", id, err)
	}
	if err := handler.Validate(value); err != nil {
		return Value{}, fmt.Errorf("validate parsed %q value: %w", id, err)
	}
	return value, nil
}

func (registry *Registry) Compare(left, right Value) (int, error) {
	if left.TypeID() != right.TypeID() {
		return 0, fmt.Errorf("cannot compare data types %q and %q", left.TypeID(), right.TypeID())
	}
	handler, err := registry.Require(left.TypeID())
	if err != nil {
		return 0, err
	}
	return handler.Compare(left, right)
}

func (registry *Registry) Render(value Value, options RenderOptions) (string, error) {
	handler, err := registry.Require(value.TypeID())
	if err != nil {
		return "", err
	}
	return handler.Render(value, options)
}

func validateDescriptor(descriptor Descriptor, version string) error {
	if strings.TrimSpace(string(descriptor.ID)) == "" {
		return fmt.Errorf("identity is required")
	}
	if descriptor.Version != version {
		return fmt.Errorf("version %q is incompatible with registry version %q", descriptor.Version, version)
	}
	if strings.TrimSpace(descriptor.Name) == "" {
		return fmt.Errorf("name is required for %q", descriptor.ID)
	}
	return nil
}

func normalizeID(value ID) ID {
	return ID(strings.ToLower(strings.TrimSpace(string(value))))
}
