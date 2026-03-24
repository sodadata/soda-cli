package lint

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v6"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"gopkg.in/yaml.v3"
)

var printer = message.NewPrinter(language.English)

// LintResult holds the outcome of validating one file.
type LintResult struct {
	File   string      `json:"file"`
	Valid  bool        `json:"valid"`
	Errors []LintError `json:"errors,omitempty"`
}

// LintError represents a single schema validation error.
type LintError struct {
	Path    string `json:"path"`
	Message string `json:"message"`
}

// LintFile validates a single YAML file against the contract schema.
func LintFile(filePath string) (*LintResult, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("could not read %s: %w", filePath, err)
	}

	result := &LintResult{File: filePath, Valid: true}

	// Parse YAML
	var doc interface{}
	if err := yaml.Unmarshal(data, &doc); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, LintError{
			Path:    "$",
			Message: fmt.Sprintf("YAML parse error: %v", err),
		})
		return result, nil
	}

	if doc == nil {
		result.Valid = false
		result.Errors = append(result.Errors, LintError{
			Path:    "$",
			Message: "file is empty",
		})
		return result, nil
	}

	// Convert YAML map keys from interface{} to string for JSON schema compatibility
	doc = normalizeYAML(doc)

	// Validate against schema
	schema, err := loadSchema()
	if err != nil {
		return nil, fmt.Errorf("could not load schema: %w", err)
	}

	if err := schema.Validate(doc); err != nil {
		result.Valid = false
		result.Errors = flattenErrors(err)
	}

	return result, nil
}

// LintFiles validates multiple YAML files and returns results for each.
func LintFiles(paths []string) ([]*LintResult, error) {
	var results []*LintResult
	for _, p := range paths {
		r, err := LintFile(p)
		if err != nil {
			return nil, err
		}
		results = append(results, r)
	}
	return results, nil
}

// ResultsJSON returns the lint results as a JSON string.
func ResultsJSON(results []*LintResult) (string, error) {
	b, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// flattenErrors converts a jsonschema.ValidationError tree into a flat list of LintErrors.
// It collects only leaf errors to avoid generic wrapper messages like "doesn't match anyOf".
func flattenErrors(err error) []LintError {
	ve, ok := err.(*jsonschema.ValidationError)
	if !ok {
		return []LintError{{Path: "$", Message: err.Error()}}
	}

	var errors []LintError
	collectLeafErrors(ve, &errors)
	return errors
}

func collectLeafErrors(ve *jsonschema.ValidationError, out *[]LintError) {
	if len(ve.Causes) == 0 {
		path := segmentsToPath(ve.InstanceLocation)
		msg := ve.ErrorKind.LocalizedString(printer)
		if msg == "" {
			msg = ve.Error()
		}
		*out = append(*out, LintError{Path: path, Message: msg})
		return
	}
	for _, cause := range ve.Causes {
		collectLeafErrors(cause, out)
	}
}

// segmentsToPath converts a []string like ["columns", "0", "checks", "0"] to "$.columns[0].checks[0]".
func segmentsToPath(segments []string) string {
	if len(segments) == 0 {
		return "$"
	}
	var sb strings.Builder
	sb.WriteString("$")
	for _, part := range segments {
		// Check if part is a number (array index)
		isIndex := true
		for _, c := range part {
			if c < '0' || c > '9' {
				isIndex = false
				break
			}
		}
		if isIndex && len(part) > 0 {
			sb.WriteString("[" + part + "]")
		} else {
			sb.WriteString("." + part)
		}
	}
	return sb.String()
}

// normalizeYAML converts map[interface{}]interface{} (from yaml.v3) into
// map[string]interface{} so that the JSON schema validator can handle it.
func normalizeYAML(v interface{}) interface{} {
	switch val := v.(type) {
	case map[string]interface{}:
		out := make(map[string]interface{}, len(val))
		for k, v := range val {
			out[k] = normalizeYAML(v)
		}
		return out
	case map[interface{}]interface{}:
		out := make(map[string]interface{}, len(val))
		for k, v := range val {
			out[fmt.Sprintf("%v", k)] = normalizeYAML(v)
		}
		return out
	case []interface{}:
		for i, item := range val {
			val[i] = normalizeYAML(item)
		}
		return val
	default:
		return v
	}
}
