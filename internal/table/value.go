// Package table defines adapter-neutral canonical data, provenance, stable row
// streams, and bounded spooling.
package table

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"regexp"
	"time"

	"github.com/cockroachdb/apd/v3"
)

type ValueKind string

const (
	KindNull       ValueKind = "null"
	KindString     ValueKind = "string"
	KindBoolean    ValueKind = "boolean"
	KindInteger    ValueKind = "integer"
	KindDecimal    ValueKind = "decimal"
	KindDate       ValueKind = "date"
	KindTime       ValueKind = "time"
	KindJSONObject ValueKind = "json_object"
	KindJSONArray  ValueKind = "json_array"
)

var (
	integerPattern = regexp.MustCompile("^-?(0|[1-9][0-9]*)$")
	numberPattern  = regexp.MustCompile("^-?(0|[1-9][0-9]*)(\\.[0-9]+)?([eE][+-]?[0-9]+)?$")
)

// Value is a tagged canonical value. text retains string content, local
// date/time text, or the exact source numeric lexeme.
type Value struct {
	kind    ValueKind
	text    string
	boolean bool
	json    []byte
}

func NullValue() Value {
	return Value{kind: KindNull}
}

func StringValue(value string) Value {
	return Value{kind: KindString, text: value}
}

func BooleanValue(value bool) Value {
	return Value{kind: KindBoolean, boolean: value}
}

func IntegerValue(lexeme string) (Value, error) {
	if !integerPattern.MatchString(lexeme) {
		return Value{}, fmt.Errorf("invalid exact integer lexeme %q", lexeme)
	}
	if _, ok := new(big.Int).SetString(lexeme, 10); !ok {
		return Value{}, fmt.Errorf("invalid exact integer lexeme %q", lexeme)
	}
	return Value{kind: KindInteger, text: lexeme}, nil
}

func DecimalValue(lexeme string) (Value, error) {
	if !numberPattern.MatchString(lexeme) {
		return Value{}, fmt.Errorf("invalid exact decimal lexeme %q", lexeme)
	}
	if _, _, err := apd.NewFromString(lexeme); err != nil {
		return Value{}, fmt.Errorf("parse exact decimal %q: %w", lexeme, err)
	}
	return Value{kind: KindDecimal, text: lexeme}, nil
}

type LocalDate struct {
	Year  int
	Month time.Month
	Day   int
}

func DateValue(year int, month time.Month, day int) (Value, error) {
	date := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
	if date.Year() != year || date.Month() != month || date.Day() != day {
		return Value{}, fmt.Errorf("invalid local date %04d-%02d-%02d", year, month, day)
	}
	return Value{kind: KindDate, text: date.Format("2006-01-02")}, nil
}

type LocalTime struct {
	Hour       int
	Minute     int
	Second     int
	Nanosecond int
}

func TimeValue(hour, minute, second, nanosecond int) (Value, error) {
	if hour < 0 || hour > 23 || minute < 0 || minute > 59 || second < 0 || second > 59 || nanosecond < 0 || nanosecond >= int(time.Second) {
		return Value{}, fmt.Errorf("invalid local time %02d:%02d:%02d.%09d", hour, minute, second, nanosecond)
	}
	value := time.Date(0, time.January, 1, hour, minute, second, nanosecond, time.UTC)
	return Value{kind: KindTime, text: value.Format("15:04:05.999999999")}, nil
}

func JSONValue(raw []byte) (Value, error) {
	if !json.Valid(raw) {
		return Value{}, errors.New("invalid canonical JSON")
	}
	var compact bytes.Buffer
	if err := json.Compact(&compact, raw); err != nil {
		return Value{}, fmt.Errorf("compact canonical JSON: %w", err)
	}
	contents := compact.Bytes()
	if len(contents) == 0 {
		return Value{}, errors.New("empty canonical JSON")
	}
	var kind ValueKind
	switch contents[0] {
	case '{':
		kind = KindJSONObject
	case '[':
		kind = KindJSONArray
	default:
		return Value{}, errors.New("canonical JSON value must be an object or array")
	}
	return Value{kind: kind, json: append([]byte(nil), contents...)}, nil
}

func (value Value) Kind() ValueKind {
	return value.kind
}

func (value Value) IsNull() bool {
	return value.kind == KindNull
}

func (value Value) StringContent() (string, bool) {
	return value.text, value.kind == KindString
}

func (value Value) Boolean() (bool, bool) {
	return value.boolean, value.kind == KindBoolean
}

func (value Value) RawNumericLexeme() (string, bool) {
	return value.text, value.kind == KindInteger || value.kind == KindDecimal
}

func (value Value) Date() (LocalDate, bool) {
	if value.kind != KindDate {
		return LocalDate{}, false
	}
	parsed, err := time.Parse("2006-01-02", value.text)
	if err != nil {
		return LocalDate{}, false
	}
	return LocalDate{Year: parsed.Year(), Month: parsed.Month(), Day: parsed.Day()}, true
}

func (value Value) Time() (LocalTime, bool) {
	if value.kind != KindTime {
		return LocalTime{}, false
	}
	parsed, err := time.Parse("15:04:05.999999999", value.text)
	if err != nil {
		return LocalTime{}, false
	}
	return LocalTime{
		Hour:       parsed.Hour(),
		Minute:     parsed.Minute(),
		Second:     parsed.Second(),
		Nanosecond: parsed.Nanosecond(),
	}, true
}

func (value Value) CanonicalJSON() ([]byte, bool) {
	if value.kind != KindJSONObject && value.kind != KindJSONArray {
		return nil, false
	}
	return append([]byte(nil), value.json...), true
}

type serializedValue struct {
	Kind  ValueKind       `json:"kind"`
	Text  string          `json:"text,omitempty"`
	Bool  *bool           `json:"boolean,omitempty"`
	Value json.RawMessage `json:"value,omitempty"`
}

func (value Value) MarshalJSON() ([]byte, error) {
	serialized := serializedValue{Kind: value.kind}
	switch value.kind {
	case KindNull:
	case KindString, KindInteger, KindDecimal, KindDate, KindTime:
		serialized.Text = value.text
	case KindBoolean:
		boolean := value.boolean
		serialized.Bool = &boolean
	case KindJSONObject, KindJSONArray:
		serialized.Value = append(json.RawMessage(nil), value.json...)
	default:
		return nil, fmt.Errorf("unsupported canonical value kind %q", value.kind)
	}
	return json.Marshal(serialized)
}

func (value *Value) UnmarshalJSON(data []byte) error {
	var serialized serializedValue
	if err := json.Unmarshal(data, &serialized); err != nil {
		return err
	}
	var decoded Value
	var err error
	switch serialized.Kind {
	case KindNull:
		decoded = NullValue()
	case KindString:
		decoded = StringValue(serialized.Text)
	case KindBoolean:
		if serialized.Bool == nil {
			return errors.New("canonical Boolean is missing boolean value")
		}
		decoded = BooleanValue(*serialized.Bool)
	case KindInteger:
		decoded, err = IntegerValue(serialized.Text)
	case KindDecimal:
		decoded, err = DecimalValue(serialized.Text)
	case KindDate:
		var parsed time.Time
		parsed, err = time.Parse("2006-01-02", serialized.Text)
		if err == nil {
			decoded, err = DateValue(parsed.Year(), parsed.Month(), parsed.Day())
		}
	case KindTime:
		var parsed time.Time
		parsed, err = time.Parse("15:04:05.999999999", serialized.Text)
		if err == nil {
			decoded, err = TimeValue(parsed.Hour(), parsed.Minute(), parsed.Second(), parsed.Nanosecond())
		}
	case KindJSONObject, KindJSONArray:
		decoded, err = JSONValue(serialized.Value)
		if err == nil && decoded.kind != serialized.Kind {
			err = fmt.Errorf("canonical JSON kind %q does not match value", serialized.Kind)
		}
	default:
		err = fmt.Errorf("unsupported canonical value kind %q", serialized.Kind)
	}
	if err != nil {
		return err
	}
	*value = decoded
	return nil
}
