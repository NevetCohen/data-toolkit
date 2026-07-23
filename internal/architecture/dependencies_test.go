package architecture_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

const internalImportPrefix = "data-toolkit/internal/"

var forbiddenDependencies = map[string]map[string]struct{}{
	"contract":     forbidden("config", "operation", "adapter", "orchestrator", "report", "application", "cli", "tui"),
	"config":       forbidden("operation", "adapter", "orchestrator", "application", "cli", "tui"),
	"table":        forbidden("config", "operation", "adapter", "orchestrator", "report", "application", "cli", "tui"),
	"report":       forbidden("operation", "adapter", "orchestrator", "application", "cli", "tui"),
	"operation":    forbidden("adapter", "orchestrator", "application", "cli", "tui"),
	"adapter":      forbidden("orchestrator", "application", "cli", "tui"),
	"orchestrator": forbidden("application", "cli", "tui"),
	"application":  forbidden("cli", "tui"),
	"cli":          forbidden("tui"),
	"tui":          forbidden("cli"),
}

func TestInternalDependencyBoundaries(t *testing.T) {
	t.Parallel()

	root := repositoryRoot(t)
	fset := token.NewFileSet()

	err := filepath.WalkDir(filepath.Join(root, "internal"), func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || filepath.Ext(path) != ".go" {
			return nil
		}

		relative, err := filepath.Rel(filepath.Join(root, "internal"), path)
		if err != nil {
			return err
		}
		parts := strings.Split(filepath.ToSlash(relative), "/")
		if len(parts) < 2 {
			return nil
		}
		sourceLayer := parts[0]
		blocked := forbiddenDependencies[sourceLayer]
		if len(blocked) == 0 {
			return nil
		}

		parsed, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if err != nil {
			return err
		}
		for _, imported := range parsed.Imports {
			importPath, err := strconv.Unquote(imported.Path.Value)
			if err != nil {
				return err
			}
			if !strings.HasPrefix(importPath, internalImportPrefix) {
				continue
			}
			target := strings.Split(strings.TrimPrefix(importPath, internalImportPrefix), "/")[0]
			if _, denied := blocked[target]; denied {
				position := fset.Position(imported.Pos())
				t.Errorf("%s layer must not import %s layer (%s)", sourceLayer, target, position)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("scan internal packages: %v", err)
	}
}

func TestOrchestratorContainsNoLogicalStreamTransformations(t *testing.T) {
	t.Parallel()

	root := filepath.Join(repositoryRoot(t), "internal", "orchestrator")
	fset := token.NewFileSet()
	forbiddenNames := map[string]struct{}{
		"concatenateStream":    {},
		"renameStream":         {},
		"newConcatenateStream": {},
		"newRenameStream":      {},
	}
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || filepath.Ext(path) != ".go" {
			return nil
		}
		parsed, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			return err
		}
		ast.Inspect(parsed, func(node ast.Node) bool {
			switch declaration := node.(type) {
			case *ast.TypeSpec:
				if _, forbidden := forbiddenNames[declaration.Name.Name]; forbidden {
					t.Errorf("orchestrator must not declare logical stream type %q (%s)", declaration.Name.Name, fset.Position(declaration.Pos()))
				}
			case *ast.FuncDecl:
				if _, forbidden := forbiddenNames[declaration.Name.Name]; forbidden {
					t.Errorf("orchestrator must not declare logical stream constructor %q (%s)", declaration.Name.Name, fset.Position(declaration.Pos()))
				}
			}
			return true
		})
		return nil
	})
	if err != nil {
		t.Fatalf("scan orchestrator package: %v", err)
	}
}

func forbidden(layers ...string) map[string]struct{} {
	result := make(map[string]struct{}, len(layers))
	for _, layer := range layers {
		result[layer] = struct{}{}
	}
	return result
}

func repositoryRoot(t *testing.T) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve architecture test path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(filename), "..", ".."))
}
