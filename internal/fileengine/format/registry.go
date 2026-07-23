package format

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
)

type Registry struct {
	version   string
	mu        sync.RWMutex
	providers map[string]Provider
}

func NewRegistry(version string) *Registry {
	return &Registry{version: version, providers: make(map[string]Provider)}
}

func (registry *Registry) RegisterConstructor(constructor Constructor) error {
	if constructor == nil {
		return fmt.Errorf("register format: constructor is nil")
	}
	provider, err := constructor()
	if err != nil {
		return fmt.Errorf("construct format: %w", err)
	}
	return registry.Register(provider)
}

func (registry *Registry) Register(provider Provider) error {
	if registry == nil {
		return fmt.Errorf("register format: registry is nil")
	}
	if provider == nil {
		return fmt.Errorf("register format: provider is nil")
	}
	descriptor := provider.Descriptor()
	if err := validateDescriptor(descriptor, provider, registry.version); err != nil {
		return fmt.Errorf("register format: %w", err)
	}
	key := normalize(descriptor.ID)
	registry.mu.Lock()
	defer registry.mu.Unlock()
	if _, exists := registry.providers[key]; exists {
		return fmt.Errorf("register format: identity %q is already registered", descriptor.ID)
	}
	registry.providers[key] = provider
	return nil
}

func (registry *Registry) Require(id string, capabilities ...Capability) (Provider, error) {
	if registry == nil {
		return nil, fmt.Errorf("format registry is nil")
	}
	key := normalize(id)
	registry.mu.RLock()
	provider, exists := registry.providers[key]
	registry.mu.RUnlock()
	if !exists {
		return nil, fmt.Errorf("format %q is not registered", id)
	}
	declared := capabilitySet(provider.Descriptor().Capabilities)
	for _, capability := range capabilities {
		if !declared[capability] {
			return nil, fmt.Errorf("format %q does not support capability %q", id, capability)
		}
	}
	return provider, nil
}

func (registry *Registry) Map(ctx context.Context, request MapRequest) (Mapping, error) {
	provider, err := registry.Require(request.FormatID, CapabilityMap)
	if err != nil {
		return Mapping{}, err
	}
	mapper, ok := provider.(Mapper)
	if !ok {
		return Mapping{}, fmt.Errorf("format %q declared map capability without mapper", request.FormatID)
	}
	return mapper.Map(ctx, request)
}

func (registry *Registry) Descriptors() []Descriptor {
	if registry == nil {
		return nil
	}
	registry.mu.RLock()
	result := make([]Descriptor, 0, len(registry.providers))
	for _, provider := range registry.providers {
		result = append(result, cloneDescriptor(provider.Descriptor()))
	}
	registry.mu.RUnlock()
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result
}

func validateDescriptor(descriptor Descriptor, provider Provider, version string) error {
	if strings.TrimSpace(descriptor.ID) == "" || strings.TrimSpace(descriptor.Name) == "" {
		return fmt.Errorf("identity and name are required")
	}
	if descriptor.Version != version {
		return fmt.Errorf("version %q is incompatible with registry version %q", descriptor.Version, version)
	}
	declared := make(map[Capability]bool, len(descriptor.Capabilities))
	for _, capability := range descriptor.Capabilities {
		if !knownCapability(capability) {
			return fmt.Errorf("format %q declares unknown capability %q", descriptor.ID, capability)
		}
		if declared[capability] {
			return fmt.Errorf("format %q repeats capability %q", descriptor.ID, capability)
		}
		declared[capability] = true
	}
	implemented := map[Capability]bool{}
	_, implemented[CapabilityInspect] = provider.(Inspector)
	_, implemented[CapabilityMap] = provider.(Mapper)
	_, implemented[CapabilityRead] = provider.(Reader)
	_, implemented[CapabilityWrite] = provider.(Writer)
	_, implemented[CapabilityStyle] = provider.(Styler)
	_, implemented[CapabilityValidate] = provider.(Validator)
	for _, capability := range allCapabilities() {
		if declared[capability] != implemented[capability] {
			return fmt.Errorf("format %q capability %q declaration does not match implemented interface", descriptor.ID, capability)
		}
	}
	return nil
}

func allCapabilities() []Capability {
	return []Capability{CapabilityInspect, CapabilityMap, CapabilityRead, CapabilityWrite, CapabilityStyle, CapabilityValidate}
}

func knownCapability(value Capability) bool {
	for _, capability := range allCapabilities() {
		if capability == value {
			return true
		}
	}
	return false
}

func capabilitySet(values []Capability) map[Capability]bool {
	result := make(map[Capability]bool, len(values))
	for _, value := range values {
		result[value] = true
	}
	return result
}

func cloneDescriptor(source Descriptor) Descriptor {
	source.Extensions = append([]string(nil), source.Extensions...)
	source.Capabilities = append([]Capability(nil), source.Capabilities...)
	sort.Slice(source.Capabilities, func(i, j int) bool { return source.Capabilities[i] < source.Capabilities[j] })
	return source
}

func normalize(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}
