package table

import (
	"errors"
	"fmt"
	"strconv"
	"time"
)

type RenderOptions struct {
	NullDisplay      string
	DecimalPrecision int
	DateFormat       string
	TimeFormat       string
}

func Render(value Value, options RenderOptions) (string, error) {
	switch value.kind {
	case KindNull:
		return options.NullDisplay, nil
	case KindString:
		return value.text, nil
	case KindBoolean:
		return strconv.FormatBool(value.boolean), nil
	case KindInteger:
		return value.text, nil
	case KindDecimal:
		return FormatNumeric(value, options.DecimalPrecision)
	case KindDate:
		date, ok := value.Date()
		if !ok {
			return "", errors.New("invalid canonical date")
		}
		return time.Date(date.Year, date.Month, date.Day, 0, 0, 0, 0, time.UTC).Format(options.DateFormat), nil
	case KindTime:
		localTime, ok := value.Time()
		if !ok {
			return "", errors.New("invalid canonical time")
		}
		return time.Date(0, time.January, 1, localTime.Hour, localTime.Minute, localTime.Second, localTime.Nanosecond, time.UTC).Format(options.TimeFormat), nil
	case KindJSONObject, KindJSONArray:
		return string(value.json), nil
	default:
		return "", fmt.Errorf("unsupported canonical value kind %q", value.kind)
	}
}
