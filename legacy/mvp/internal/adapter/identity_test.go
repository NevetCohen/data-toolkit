package adapter

import (
	"path/filepath"
	"testing"

	"data-toolkit/internal/contract"
)

func TestLocalOutputCannotResolveToSourceUnderAnyCollisionPolicy(t *testing.T) {
	path := filepath.Join(t.TempDir(), "source.csv")
	source, err := LocalIdentity(path)
	if err != nil {
		t.Fatal(err)
	}
	output, err := LocalIdentity(filepath.Join(filepath.Dir(path), ".", "source.csv"))
	if err != nil {
		t.Fatal(err)
	}
	for _, action := range []contract.CollisionAction{
		contract.CollisionBlock,
		contract.CollisionOverwrite,
		contract.CollisionAlternateName,
	} {
		t.Run(string(action), func(t *testing.T) {
			if err := ValidateDistinctOutputIdentities([]ResourceIdentity{source}, []ResourceIdentity{output}); err == nil {
				t.Fatal("source identity was accepted as an output")
			}
		})
	}
}

func TestGoogleOutputCannotResolveToSource(t *testing.T) {
	source, _ := GoogleIdentity("sheet-file-id")
	output, _ := GoogleIdentity("sheet-file-id")
	if err := ValidateDistinctOutputIdentities([]ResourceIdentity{source}, []ResourceIdentity{output}); err == nil {
		t.Fatal("Google source identity was accepted as an output")
	}
}
