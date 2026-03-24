package lint

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

//go:embed schema/soda_data_contract_json_schema_1_0_0.json
var embeddedSchema string

var (
	compiledSchema *jsonschema.Schema
	schemaOnce     sync.Once
	schemaErr      error
)

// loadSchema returns the compiled JSON schema, using a cached version from
// ~/.soda/cache/ if available, otherwise falling back to the embedded schema.
func loadSchema() (*jsonschema.Schema, error) {
	schemaOnce.Do(func() {
		schemaSource := embeddedSchema

		// Check for cached (newer) schema
		if cached, err := loadCachedSchema(); err == nil && len(cached) > 0 {
			schemaSource = string(cached)
		}

		compiledSchema, schemaErr = compileSchema(schemaSource)
	})
	return compiledSchema, schemaErr
}

func compileSchema(source string) (*jsonschema.Schema, error) {
	doc, err := jsonschema.UnmarshalJSON(strings.NewReader(source))
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON schema: %w", err)
	}

	c := jsonschema.NewCompiler()
	if err := c.AddResource("schema.json", doc); err != nil {
		return nil, fmt.Errorf("failed to add schema resource: %w", err)
	}

	sch, err := c.Compile("schema.json")
	if err != nil {
		return nil, fmt.Errorf("failed to compile schema: %w", err)
	}

	return sch, nil
}

func loadCachedSchema() ([]byte, error) {
	path, err := CachedSchemaPath()
	if err != nil {
		return nil, err
	}
	return os.ReadFile(path)
}

// CachedSchemaPath returns the path to the cached schema file (~/.soda/cache/contract_schema.json).
func CachedSchemaPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".soda", "cache", "contract_schema.json"), nil
}

// UpdateCachedSchema writes new schema bytes to the cache location.
func UpdateCachedSchema(data []byte) error {
	path, err := CachedSchemaPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
