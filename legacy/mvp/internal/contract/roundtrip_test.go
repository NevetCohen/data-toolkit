package contract

import (
	"bytes"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestWorkflowFixturesRoundTripInBothFormats(t *testing.T) {
	for _, format := range []DocumentFormat{DocumentJSON, DocumentYAML} {
		t.Run(string(format), func(t *testing.T) {
			extension := string(format)
			contents, err := os.ReadFile("testdata/workflow." + extension)
			if err != nil {
				t.Fatal(err)
			}
			before, err := DecodeWorkflow(bytes.NewReader(contents), format)
			if err != nil {
				t.Fatal(err)
			}

			var encoded bytes.Buffer
			if err := EncodeWorkflow(&encoded, before, format); err != nil {
				t.Fatal(err)
			}
			after, err := DecodeWorkflow(&encoded, format)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(before, after) {
				t.Fatalf("round trip changed workflow:\nwant %#v\ngot  %#v", before, after)
			}
		})
	}
}

func TestRejectionFixturesAreExplicit(t *testing.T) {
	tests := []struct {
		path       string
		format     DocumentFormat
		wantedPath string
	}{
		{path: "testdata/reject/ambiguous-expression.yaml", format: DocumentYAML, wantedPath: "$.operations[0].condition"},
		{path: "testdata/reject/incomplete-operation.json", format: DocumentJSON, wantedPath: "$.operations[0].condition"},
	}

	for _, test := range tests {
		t.Run(test.path, func(t *testing.T) {
			contents, err := os.ReadFile(test.path)
			if err != nil {
				t.Fatal(err)
			}
			_, err = DecodeWorkflow(bytes.NewReader(contents), test.format)
			if err == nil {
				t.Fatal("expected rejection fixture to fail")
			}
			if !strings.Contains(err.Error(), test.wantedPath) {
				t.Fatalf("error %q does not contain path %q", err, test.wantedPath)
			}
		})
	}
}
