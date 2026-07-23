package architecture_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestV1DoesNotImportLegacyOrOutwardLayers(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	rules := map[string][]string{
		"internal/core":         {"internal/logical", "internal/fileengine", "internal/orchestrator", "internal/application", "internal/interfaces"},
		"internal/logical":      {"internal/fileengine", "internal/orchestrator", "internal/application", "internal/interfaces"},
		"internal/fileengine":   {"internal/logical", "internal/orchestrator", "internal/application", "internal/interfaces"},
		"internal/orchestrator": {"internal/application", "internal/interfaces"},
		"internal/application":  {"internal/interfaces"},
	}
	for layer, forbidden := range rules {
		layerPath := filepath.Join(root, filepath.FromSlash(layer))
		err := filepath.WalkDir(layerPath, func(path string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
				return nil
			}
			file, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
			if err != nil {
				return err
			}
			for _, declaration := range file.Imports {
				importPath, err := strconv.Unquote(declaration.Path.Value)
				if err != nil {
					return err
				}
				if strings.Contains(importPath, "/legacy/mvp") {
					t.Errorf("%s imports isolated legacy module %q", path, importPath)
				}
				for _, denied := range forbidden {
					if strings.Contains(importPath, "/"+denied) {
						t.Errorf("%s violates %s boundary with import %q", path, layer, importPath)
					}
				}
			}
			return nil
		})
		if err != nil {
			t.Fatalf("scan %s: %v", layer, err)
		}
	}

	for _, relative := range []string{"cmd", "internal"} {
		err := filepath.WalkDir(filepath.Join(root, relative), func(path string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") {
				return nil
			}
			file, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
			if err != nil {
				return err
			}
			ast.Inspect(file, func(node ast.Node) bool {
				declaration, ok := node.(*ast.ImportSpec)
				if ok && strings.Contains(declaration.Path.Value, "legacy/mvp") {
					t.Errorf("%s imports legacy code", path)
				}
				return true
			})
			return nil
		})
		if err != nil {
			t.Fatalf("scan legacy imports: %v", err)
		}
	}
}
