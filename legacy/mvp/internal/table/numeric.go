package table

import (
	"errors"
	"fmt"
	"strings"

	"github.com/cockroachdb/apd/v3"
)

func CompareNumeric(left, right Value) (int, error) {
	leftDecimal, err := numericDecimal(left)
	if err != nil {
		return 0, fmt.Errorf("left value: %w", err)
	}
	rightDecimal, err := numericDecimal(right)
	if err != nil {
		return 0, fmt.Errorf("right value: %w", err)
	}
	return leftDecimal.Cmp(rightDecimal), nil
}

func FormatNumeric(value Value, decimalPlaces int) (string, error) {
	if decimalPlaces < 0 {
		return "", errors.New("decimal places must not be negative")
	}
	decimal, err := numericDecimal(value)
	if err != nil {
		return "", err
	}
	precision := len(strings.TrimLeft(value.text, "-+0.")) + decimalPlaces + 8
	if precision < 34 {
		precision = 34
	}
	context := apd.BaseContext.WithPrecision(uint32(precision))
	context.Rounding = apd.RoundHalfUp
	var rounded apd.Decimal
	if _, err := context.Quantize(&rounded, decimal, -int32(decimalPlaces)); err != nil {
		return "", fmt.Errorf("format exact number: %w", err)
	}
	return rounded.Text('f'), nil
}

func numericDecimal(value Value) (*apd.Decimal, error) {
	if value.kind != KindInteger && value.kind != KindDecimal {
		return nil, fmt.Errorf("canonical kind %q is not numeric", value.kind)
	}
	decimal, _, err := apd.NewFromString(value.text)
	if err != nil {
		return nil, fmt.Errorf("parse exact numeric lexeme %q: %w", value.text, err)
	}
	return decimal, nil
}
