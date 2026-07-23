package cli_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"data-toolkit/configs"
	"data-toolkit/internal/application"
	"data-toolkit/internal/config"
	"data-toolkit/internal/core/datatype"
	fileformat "data-toolkit/internal/fileengine/format"
	"data-toolkit/internal/interfaces/cli"
	logicaloperation "data-toolkit/internal/logical/operation"
)

func testService(t *testing.T) application.API {
	t.Helper()
	dataTypes, err := datatype.NewBuiltinRegistry()
	if err != nil {
		t.Fatal(err)
	}
	defaults, err := config.LoadDefaults()
	if err != nil {
		t.Fatal(err)
	}
	service, err := application.New(application.Options{
		DataTypes:  dataTypes,
		Operations: logicaloperation.NewRegistry(logicaloperation.CurrentVersion),
		Formats:    fileformat.NewRegistry(fileformat.CurrentVersion),
		Defaults:   defaults,
	})
	if err != nil {
		t.Fatal(err)
	}
	return service
}

func TestCLIUsesApplicationAPIForCapabilitiesAndConfig(t *testing.T) {
	api := testService(t)
	var output bytes.Buffer
	if err := cli.Execute(context.Background(), api, []string{"capabilities"}, &output); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), `"data_types"`) || !strings.Contains(output.String(), `"file_formats"`) {
		t.Fatalf("capability output = %s", output.String())
	}

	payload, err := configs.Default()
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, payload, 0o600); err != nil {
		t.Fatal(err)
	}
	output.Reset()
	if err := cli.Execute(context.Background(), api, []string{"config", "validate", path}, &output); err != nil {
		t.Fatal(err)
	}
	if got := output.String(); got != "valid=true schema_version=v1\n" {
		t.Fatalf("config output = %q", got)
	}
}
