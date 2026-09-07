package table

import (
	"context"
	"errors"
	"io"
	"testing"
)

func TestSliceStreamNextHonorsCancelledContext(t *testing.T) {
	stream := NewSliceStream([]Row{{ID: "row-1"}})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	got, err := stream.Next(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Next() error = %v, want context.Canceled", err)
	}
	if got.ID != "" || got.SourceID != "" || got.SheetID != nil || got.Ordinal != 0 || len(got.Cells) != 0 {
		t.Fatalf("Next() row = %#v, want zero Row", got)
	}
}

func TestSliceStreamNextReturnsEOFAfterLastRow(t *testing.T) {
	stream := NewSliceStream([]Row{{ID: "row-1"}})

	if _, err := stream.Next(context.Background()); err != nil {
		t.Fatalf("first Next() error = %v", err)
	}
	for call := 1; call <= 2; call++ {
		if _, err := stream.Next(context.Background()); err != io.EOF {
			t.Fatalf("EOF Next() call %d error = %v, want io.EOF", call, err)
		}
	}
}

func TestSliceStreamCloseIsIdempotent(t *testing.T) {
	stream := NewSliceStream(nil)

	for call := 1; call <= 2; call++ {
		if err := stream.Close(); err != nil {
			t.Fatalf("Close() call %d error = %v, want nil", call, err)
		}
	}
}
