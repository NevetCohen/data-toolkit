package table

import "testing"

func TestJSONProjectionDoesNotFlattenUndeclaredPaths(t *testing.T) {
	document := []byte("{\"person\":{\"name\":\"נועה\",\"contact\":{\"email\":\"noa@example.com\"}},\"secret\":\"ignore\",\"items\":[1,{\"emoji\":\"🧭\"}]}")
	projected, err := ProjectJSON(document, []JSONProjection{
		{Name: "name", Pointer: "/person/name"},
		{Name: "contact", Pointer: "/person/contact"},
		{Name: "emoji", Pointer: "/items/1/emoji"},
		{Name: "missing", Pointer: "/person/phone"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(projected) != 4 {
		t.Fatalf("projection count = %d", len(projected))
	}
	if got, _ := projected[0].Value.StringContent(); got != "נועה" {
		t.Fatalf("name = %q", got)
	}
	if projected[1].Value.Kind() != KindJSONObject {
		t.Fatalf("nested object kind = %q", projected[1].Value.Kind())
	}
	if got, _ := projected[2].Value.StringContent(); got != "🧭" {
		t.Fatalf("emoji = %q", got)
	}
	if projected[3].Found || !projected[3].Value.IsNull() {
		t.Fatalf("missing path = %#v", projected[3])
	}
	for _, value := range projected {
		if value.Name == "secret" {
			t.Fatal("undeclared path was silently flattened")
		}
	}
}
