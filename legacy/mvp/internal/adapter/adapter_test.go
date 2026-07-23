package adapter

import (
	"context"
	"testing"

	"data-toolkit/internal/contract"
	"data-toolkit/internal/table"
)

func TestRegistryRejectsUnsupportedCapabilityBeforeRowsAreRead(t *testing.T) {
	fake := &fakeAdapter{
		format: contract.FormatCSV,
		capabilities: Capabilities{
			CapabilityInspect: true,
			CapabilityRead:    true,
		},
	}
	registry := NewRegistry()
	if err := registry.Register(fake); err != nil {
		t.Fatal(err)
	}
	if _, err := registry.Require(contract.FormatCSV, CapabilityWrite); err == nil {
		t.Fatal("expected unsupported write capability")
	}
	if fake.openRowsCalls != 0 {
		t.Fatalf("row reader was called %d times during capability preflight", fake.openRowsCalls)
	}
	if _, err := registry.Require(contract.FormatCSV, CapabilityInspect, CapabilityRead); err != nil {
		t.Fatal(err)
	}
}

type fakeAdapter struct {
	format        contract.DataFormat
	capabilities  Capabilities
	openRowsCalls int
}

func (adapter *fakeAdapter) Format() contract.DataFormat {
	return adapter.format
}

func (adapter *fakeAdapter) Capabilities() Capabilities {
	copy := make(Capabilities, len(adapter.capabilities))
	for capability, supported := range adapter.capabilities {
		copy[capability] = supported
	}
	return copy
}

func (adapter *fakeAdapter) Inspect(context.Context, SourceRequest) (Inspection, error) {
	return Inspection{}, nil
}

func (adapter *fakeAdapter) Map(context.Context, SourceRequest, MappingMode) (MappingReport, error) {
	return MappingReport{}, nil
}

func (adapter *fakeAdapter) OpenRows(context.Context, SourceRequest, table.Table) (table.RowStream, error) {
	adapter.openRowsCalls++
	return table.NewSliceRowStream(nil), nil
}

func (adapter *fakeAdapter) NewWriter(context.Context, OutputRequest, table.Table) (RowWriter, error) {
	return nil, nil
}

func (adapter *fakeAdapter) ValidateOutput(context.Context, OutputRequest) error {
	return nil
}
