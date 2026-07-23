package config

import (
	"testing"
	"time"

	"data-toolkit/internal/contract"
)

func TestApprovedBuiltInDefaults(t *testing.T) {
	defaults := BuiltInDefaults("temporary")
	if defaults.NullDisplay != "-" {
		t.Fatalf("null display = %q", defaults.NullDisplay)
	}
	if defaults.DecimalPrecision != 3 {
		t.Fatalf("decimal precision = %d", defaults.DecimalPrecision)
	}
	date := time.Date(2026, time.February, 2, 14, 34, 0, 0, time.UTC)
	if got := date.Format(defaults.DateFormat); got != "02/02/26" {
		t.Fatalf("date display = %q", got)
	}
	if got := date.Format(defaults.TimeFormat); got != "14:34" {
		t.Fatalf("time display = %q", got)
	}
	if defaults.PhoneFormat != "052-6105412" {
		t.Fatalf("phone display example = %q", defaults.PhoneFormat)
	}
	if defaults.TextEncoding != "UTF-8" {
		t.Fatalf("text encoding = %q", defaults.TextEncoding)
	}
	if defaults.DeduplicateKeep != contract.KeepFirst {
		t.Fatalf("deduplicate keep = %q", defaults.DeduplicateKeep)
	}
	if defaults.NotScope != contract.NotScopeBoth {
		t.Fatalf("not scope = %q", defaults.NotScope)
	}
}
