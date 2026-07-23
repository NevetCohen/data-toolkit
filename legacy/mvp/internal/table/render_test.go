package table

import "testing"

func TestNullRenderingDoesNotConflateCanonicalHyphen(t *testing.T) {
	options := RenderOptions{
		NullDisplay: "-", DecimalPrecision: 3, DateFormat: "02/01/06", TimeFormat: "15:04",
	}
	null := NullValue()
	hyphen := StringValue("-")
	nullRendered, err := Render(null, options)
	if err != nil {
		t.Fatal(err)
	}
	hyphenRendered, err := Render(hyphen, options)
	if err != nil {
		t.Fatal(err)
	}
	if nullRendered != "-" || hyphenRendered != "-" {
		t.Fatalf("rendered values = %q and %q", nullRendered, hyphenRendered)
	}
	if !null.IsNull() || hyphen.IsNull() || null.Kind() == hyphen.Kind() {
		t.Fatal("canonical null and literal hyphen were conflated")
	}
}
