// Package configs embeds the authoritative Data Toolkit V1 default
// configuration and its schema.
package configs

import (
	"embed"
	"fmt"
)

//go:embed default.yaml schema/config-v1.schema.json
var files embed.FS

func Default() ([]byte, error) {
	return readCopy("default.yaml")
}

func Schema() ([]byte, error) {
	return readCopy("schema/config-v1.schema.json")
}

func readCopy(name string) ([]byte, error) {
	contents, err := files.ReadFile(name)
	if err != nil {
		return nil, fmt.Errorf("read embedded config asset %q: %w", name, err)
	}
	return append([]byte(nil), contents...), nil
}
