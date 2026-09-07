package datatype

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/big"
	"regexp"
	"strconv"
	"strings"
	"time"

	"data-toolkit/internal/core"

	"github.com/cockroachdb/apd/v3"
)

const (
	NullID    ID = "null"
	StringID  ID = "string"
	BooleanID ID = "boolean"
	IntegerID ID = "integer"
	DecimalID ID = "decimal"
	DateID    ID = "date"
	TimeID    ID = "time"
)

var (
	integerPattern = regexp.MustCompile(`^-?(0|[1-9][0-9]*)$`)
	decimalPattern = regexp.MustCompile(`^-?(0|[1-9][0-9]*)(\.[0-9]+)?([eE][+-]?[0-9]+)?$`)
)

type builtinHandler struct {
	descriptor DataTypeDescriptor
	parse      func(string) ([]byte, error)
	validate   func([]byte) error
	compare    func([]byte, []byte) (int, error)
	render     func([]byte, RenderOptions) (string, error)
}

func (handler builtinHandler) Descriptor() Descriptor {
	return handler.descriptor
}

func (handler builtinHandler) Parse(input string) (core.Value, error) {
	encoded, err := handler.parse(input)
	if err != nil {
		return core.Value{}, err
	}
	return core.NewValue(handler.descriptor.ID, encoded)
}

func (handler builtinHandler) Validate(value core.Value) error {
	if value.TypeID() != handler.descriptor.ID {
		return fmt.Errorf("value type %q does not match %q", value.TypeID(), handler.descriptor.ID)
	}
	return handler.validate(value.Encoded())
}

func (handler builtinHandler) Compare(left, right core.Value) (int, error) {
	if err := handler.Validate(left); err != nil {
		return 0, err
	}
	if err := handler.Validate(right); err != nil {
		return 0, err
	}
	return handler.compare(left.Encoded(), right.Encoded())
}

func (handler builtinHandler) Render(value core.Value, options RenderOptions) (string, error) {
	if err := handler.Validate(value); err != nil {
		return "", err
	}
	return handler.render(value.Encoded(), options)
}

func NewBuiltinRegistry() (*Registry, error) {
	registry := NewRegistry(CurrentVersion)
	for _, handler := range builtinHandlers() {
		current := handler
		if err := registry.RegisterConstructor(func() (Handler, error) { return current, nil }); err != nil {
			return nil, err
		}
	}
	return registry, nil
}

func builtinHandlers() []Handler {
	rawRender := func(encoded []byte, _ RenderOptions) (string, error) {
		return string(encoded), nil
	}
	return []Handler{
		builtinHandler{
			descriptor: DataTypeDescriptor{ID: NullID, Version: CurrentVersion, Name: "Null", CanonicalKind: CanonicalKindNull},
			parse: func(input string) ([]byte, error) {
				if input != "null" {
					return nil, fmt.Errorf("null input must be literal null")
				}
				return []byte("null"), nil
			},
			validate: func(encoded []byte) error {
				if string(encoded) != "null" {
					return fmt.Errorf("invalid canonical null")
				}
				return nil
			},
			compare: func(_, _ []byte) (int, error) { return 0, nil },
			render: func(_ []byte, options RenderOptions) (string, error) {
				return options.NullDisplay, nil
			},
		},
		builtinHandler{
			descriptor: DataTypeDescriptor{ID: StringID, Version: CurrentVersion, Name: "String", CanonicalKind: CanonicalKindString},
			parse: func(input string) ([]byte, error) {
				if input == "" {
					return nil, nil
				}
				return json.Marshal(input)
			},
			validate: func(encoded []byte) error {
				if len(encoded) == 0 {
					return nil
				}
				var value string
				if err := json.Unmarshal(encoded, &value); err != nil {
					return fmt.Errorf("invalid canonical string: %w", err)
				}
				return nil
			},
			compare: func(left, right []byte) (int, error) {
				if len(left) == 0 || len(right) == 0 {
					return strings.Compare(string(left), string(right)), nil
				}
				var leftValue, rightValue string
				if err := json.Unmarshal(left, &leftValue); err != nil {
					return 0, err
				}
				if err := json.Unmarshal(right, &rightValue); err != nil {
					return 0, err
				}
				return strings.Compare(leftValue, rightValue), nil
			},
			render: func(encoded []byte, _ RenderOptions) (string, error) {
				if len(encoded) == 0 {
					return "", nil
				}
				var value string
				if err := json.Unmarshal(encoded, &value); err != nil {
					return "", err
				}
				return value, nil
			},
		},
		builtinHandler{
			descriptor: DataTypeDescriptor{ID: BooleanID, Version: CurrentVersion, Name: "Boolean", CanonicalKind: CanonicalKindBoolean},
			parse: func(input string) ([]byte, error) {
				value, err := strconv.ParseBool(input)
				if err != nil {
					return nil, err
				}
				return json.Marshal(value)
			},
			validate: func(encoded []byte) error {
				var value bool
				return json.Unmarshal(encoded, &value)
			},
			compare: func(left, right []byte) (int, error) {
				if bytes.Equal(left, right) {
					return 0, nil
				}
				if string(left) == "false" {
					return -1, nil
				}
				return 1, nil
			},
			render: rawRender,
		},
		builtinHandler{
			descriptor: DataTypeDescriptor{ID: IntegerID, Version: CurrentVersion, Name: "Integer", CanonicalKind: CanonicalKindInteger},
			parse:      validateInteger,
			validate: func(encoded []byte) error {
				_, err := validateInteger(string(encoded))
				return err
			},
			compare: compareIntegers,
			render:  rawRender,
		},
		builtinHandler{
			descriptor: DataTypeDescriptor{ID: DecimalID, Version: CurrentVersion, Name: "Decimal", CanonicalKind: CanonicalKindDecimal},
			parse:      validateDecimal,
			validate: func(encoded []byte) error {
				_, err := validateDecimal(string(encoded))
				return err
			},
			compare: compareDecimals,
			render:  rawRender,
		},
		newTemporalHandler(DateID, "Date", "2006-01-02", func(options RenderOptions) string { return options.DateFormat }),
		newTemporalHandler(TimeID, "Time", "15:04:05.999999999", func(options RenderOptions) string { return options.TimeFormat }),
	}
}

func validateInteger(input string) ([]byte, error) {
	if !integerPattern.MatchString(input) {
		return nil, fmt.Errorf("invalid exact integer %q", input)
	}
	if _, ok := new(big.Int).SetString(input, 10); !ok {
		return nil, fmt.Errorf("invalid exact integer %q", input)
	}
	return []byte(input), nil
}

func compareIntegers(left, right []byte) (int, error) {
	leftValue, leftOK := new(big.Int).SetString(string(left), 10)
	rightValue, rightOK := new(big.Int).SetString(string(right), 10)
	if !leftOK || !rightOK {
		return 0, fmt.Errorf("invalid canonical integer comparison")
	}
	return leftValue.Cmp(rightValue), nil
}

func validateDecimal(input string) ([]byte, error) {
	if !decimalPattern.MatchString(input) {
		return nil, fmt.Errorf("invalid exact decimal %q", input)
	}
	if _, _, err := apd.NewFromString(input); err != nil {
		return nil, fmt.Errorf("invalid exact decimal %q: %w", input, err)
	}
	return []byte(input), nil
}

func compareDecimals(left, right []byte) (int, error) {
	leftValue, _, err := apd.NewFromString(string(left))
	if err != nil {
		return 0, err
	}
	rightValue, _, err := apd.NewFromString(string(right))
	if err != nil {
		return 0, err
	}
	return leftValue.Cmp(rightValue), nil
}

func newTemporalHandler(id ID, name, canonicalLayout string, outputLayout func(RenderOptions) string) Handler {
	parse := func(input string) ([]byte, error) {
		value, err := time.Parse(canonicalLayout, input)
		if err != nil {
			return nil, err
		}
		return []byte(value.Format(canonicalLayout)), nil
	}
	return builtinHandler{
		descriptor: DataTypeDescriptor{ID: id, Version: CurrentVersion, Name: name, CanonicalKind: CanonicalKind(id)},
		parse:      parse,
		validate: func(encoded []byte) error {
			_, err := parse(string(encoded))
			return err
		},
		compare: func(left, right []byte) (int, error) {
			leftValue, err := time.Parse(canonicalLayout, string(left))
			if err != nil {
				return 0, err
			}
			rightValue, err := time.Parse(canonicalLayout, string(right))
			if err != nil {
				return 0, err
			}
			if leftValue.Before(rightValue) {
				return -1, nil
			}
			if leftValue.After(rightValue) {
				return 1, nil
			}
			return 0, nil
		},
		render: func(encoded []byte, options RenderOptions) (string, error) {
			value, err := time.Parse(canonicalLayout, string(encoded))
			if err != nil {
				return "", err
			}
			layout := outputLayout(options)
			if layout == "" {
				layout = canonicalLayout
			}
			return value.Format(layout), nil
		},
	}
}
