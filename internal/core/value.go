package core

import (
	"bytes"
	"fmt"
)

const stringDataTypeID DataTypeID = "string"

// Value holds an immutable, type-owned canonical representation.
type Value struct {
	typeID  DataTypeID
	encoded []byte
}

// NewValue copies encoded so callers cannot mutate the returned Value.
func NewValue(typeID DataTypeID, encoded []byte) (Value, error) {
	if err := typeID.Validate(); err != nil {
		return Value{}, fmt.Errorf("value type ID: %w", err)
	}
	if len(encoded) == 0 && typeID != stringDataTypeID {
		return Value{}, fmt.Errorf("value encoded bytes may be empty only for string")
	}

	return Value{
		typeID:  typeID,
		encoded: append([]byte(nil), encoded...),
	}, nil
}

func (value Value) TypeID() DataTypeID {
	return value.typeID
}

// Encoded returns a copy so callers cannot mutate Value.
func (value Value) Encoded() []byte {
	return append([]byte(nil), value.encoded...)
}

func (value Value) Equal(other Value) bool {
	return value.typeID == other.typeID && bytes.Equal(value.encoded, other.encoded)
}
