package table

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type JSONProjection struct {
	Name    string
	Pointer string
}

type ProjectedJSONValue struct {
	Name    string
	Pointer string
	Value   Value
	Found   bool
}

// ProjectJSON resolves only declared RFC 6901 JSON pointers. Missing paths
// become canonical null; objects and arrays remain canonical JSON values.
func ProjectJSON(document []byte, projections []JSONProjection) ([]ProjectedJSONValue, error) {
	decoder := json.NewDecoder(bytes.NewReader(document))
	decoder.UseNumber()
	var root any
	if err := decoder.Decode(&root); err != nil {
		return nil, fmt.Errorf("decode JSON source: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			return nil, errors.New("multiple JSON values are not allowed")
		}
		return nil, fmt.Errorf("decode trailing JSON source: %w", err)
	}

	result := make([]ProjectedJSONValue, 0, len(projections))
	for index, projection := range projections {
		if projection.Name == "" {
			return nil, fmt.Errorf("projections[%d].name: value is required", index)
		}
		selected, found, err := resolveJSONPointer(root, projection.Pointer)
		if err != nil {
			return nil, fmt.Errorf("projections[%d].pointer: %w", index, err)
		}
		value := NullValue()
		if found {
			value, err = canonicalJSONScalar(selected)
			if err != nil {
				return nil, fmt.Errorf("projections[%d]: %w", index, err)
			}
		}
		result = append(result, ProjectedJSONValue{
			Name: projection.Name, Pointer: projection.Pointer, Value: value, Found: found,
		})
	}
	return result, nil
}

func resolveJSONPointer(root any, pointer string) (any, bool, error) {
	if pointer == "" {
		return root, true, nil
	}
	if !strings.HasPrefix(pointer, "/") {
		return nil, false, errors.New("JSON pointer must be empty or start with '/'")
	}
	current := root
	for _, encodedSegment := range strings.Split(pointer[1:], "/") {
		segment, err := unescapeJSONPointerSegment(encodedSegment)
		if err != nil {
			return nil, false, err
		}
		switch typed := current.(type) {
		case map[string]any:
			next, exists := typed[segment]
			if !exists {
				return nil, false, nil
			}
			current = next
		case []any:
			index, err := strconv.Atoi(segment)
			if err != nil || index < 0 || index >= len(typed) {
				return nil, false, nil
			}
			current = typed[index]
		default:
			return nil, false, nil
		}
	}
	return current, true, nil
}

func unescapeJSONPointerSegment(segment string) (string, error) {
	var result strings.Builder
	for index := 0; index < len(segment); index++ {
		if segment[index] != '~' {
			result.WriteByte(segment[index])
			continue
		}
		if index+1 >= len(segment) {
			return "", errors.New("invalid JSON pointer escape")
		}
		index++
		switch segment[index] {
		case '0':
			result.WriteByte('~')
		case '1':
			result.WriteByte('/')
		default:
			return "", errors.New("invalid JSON pointer escape")
		}
	}
	return result.String(), nil
}

func canonicalJSONScalar(value any) (Value, error) {
	switch typed := value.(type) {
	case nil:
		return NullValue(), nil
	case string:
		return StringValue(typed), nil
	case bool:
		return BooleanValue(typed), nil
	case json.Number:
		lexeme := typed.String()
		if !strings.ContainsAny(lexeme, ".eE") {
			return IntegerValue(lexeme)
		}
		return DecimalValue(lexeme)
	case map[string]any, []any:
		encoded, err := json.Marshal(typed)
		if err != nil {
			return Value{}, fmt.Errorf("preserve nested canonical JSON: %w", err)
		}
		return JSONValue(encoded)
	default:
		return Value{}, fmt.Errorf("unsupported projected JSON type %T", value)
	}
}
