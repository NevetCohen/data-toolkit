package tui_test

import (
	"context"
	"testing"

	"data-toolkit/configs"
	"data-toolkit/internal/application"
	"data-toolkit/internal/config"
	"data-toolkit/internal/core/datatype"
	fileformat "data-toolkit/internal/fileengine/format"
	"data-toolkit/internal/interfaces/tui"
	logicaloperation "data-toolkit/internal/logical/operation"
)

func TestTUIClientDelegatesToApplicationAPI(t *testing.T) {
	dataTypes, _ := datatype.NewBuiltinRegistry()
	defaults, _ := config.LoadDefaults()
	service, err := application.New(application.Options{
		DataTypes: dataTypes, Operations: logicaloperation.NewRegistry(logicaloperation.CurrentVersion),
		Formats: fileformat.NewRegistry(fileformat.CurrentVersion), Defaults: defaults,
	})
	if err != nil {
		t.Fatal(err)
	}
	client := tui.New(service)
	capabilities, err := client.Capabilities(context.Background())
	if err != nil || len(capabilities.DataTypes) != 9 {
		t.Fatalf("capabilities = %#v, %v", capabilities, err)
	}
	payload, _ := configs.Default()
	result, err := client.ValidateConfig(context.Background(), payload, config.FormatYAML)
	if err != nil || !result.Valid {
		t.Fatalf("config validation = %#v, %v", result, err)
	}
}
