package metadata

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestParse_ValidJSON(t *testing.T) {
	valid := []byte(`{
		"spec": "squash-tree/v1",
		"type": "squash",
		"root": "abc123",
		"base": "def456",
		"children": [
			{"hash": "c1", "order": 1},
			{"hash": "c2", "order": 2}
		],
		"created_at": "2026-01-27T14:30:00Z",
		"strategy": "rebase"
	}`)
	meta, err := Parse(valid)
	if err != nil {
		t.Fatalf("Parse(valid): %v", err)
	}
	if meta.Spec != SpecVersionV1 {
		t.Errorf("Spec: got %q, want %q", meta.Spec, SpecVersionV1)
	}
	if meta.Type != TypeSquash {
		t.Errorf("Type: got %q, want %q", meta.Type, TypeSquash)
	}
	if meta.Root != "abc123" || meta.Base != "def456" {
		t.Errorf("Root/Base: got %q / %q", meta.Root, meta.Base)
	}
	if len(meta.Children) != 2 {
		t.Fatalf("len(Children): got %d, want 2", len(meta.Children))
	}
	if meta.Children[0].Hash != "c1" || meta.Children[0].Order != 1 {
		t.Errorf("Children[0]: got %+v", meta.Children[0])
	}
	if meta.Children[1].Hash != "c2" || meta.Children[1].Order != 2 {
		t.Errorf("Children[1]: got %+v", meta.Children[1])
	}

	// Round-trip
	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	meta2, err := Parse(data)
	if err != nil {
		t.Fatalf("Parse(round-trip): %v", err)
	}
	if meta2.Root != meta.Root || meta2.Base != meta.Base {
		t.Errorf("Round-trip: Root/Base mismatch")
	}
}

func TestParse_InvalidJSON(t *testing.T) {
	_, err := Parse([]byte("not json"))
	if err == nil {
		t.Fatal("Parse(invalid): expected error")
	}
}

func TestParse_ValidationErrors(t *testing.T) {
	tests := []struct {
		name string
		json string
		want string
	}{
		{"missing spec", `{"type":"squash","root":"r","base":"b","children":[{"hash":"c","order":1}],"created_at":"2026-01-01T00:00:00Z"}`, "spec"},
		{"wrong spec", `{"spec":"v0","type":"squash","root":"r","base":"b","children":[{"hash":"c","order":1}],"created_at":"2026-01-01T00:00:00Z"}`, "unsupported spec version"},
		{"missing type", `{"spec":"squash-tree/v1","root":"r","base":"b","children":[{"hash":"c","order":1}],"created_at":"2026-01-01T00:00:00Z"}`, "type"},
		{"wrong type", `{"spec":"squash-tree/v1","type":"other","root":"r","base":"b","children":[{"hash":"c","order":1}],"created_at":"2026-01-01T00:00:00Z"}`, "unsupported type"},
		{"missing root", `{"spec":"squash-tree/v1","type":"squash","base":"b","children":[{"hash":"c","order":1}],"created_at":"2026-01-01T00:00:00Z"}`, "root"},
		{"missing base", `{"spec":"squash-tree/v1","type":"squash","root":"r","children":[{"hash":"c","order":1}],"created_at":"2026-01-01T00:00:00Z"}`, "base"},
		{"empty children", `{"spec":"squash-tree/v1","type":"squash","root":"r","base":"b","children":[],"created_at":"2026-01-01T00:00:00Z"}`, "at least one child"},
		{"child missing hash", `{"spec":"squash-tree/v1","type":"squash","root":"r","base":"b","children":[{"hash":"","order":1}],"created_at":"2026-01-01T00:00:00Z"}`, "missing hash"},
		{"child invalid order", `{"spec":"squash-tree/v1","type":"squash","root":"r","base":"b","children":[{"hash":"c","order":0}],"created_at":"2026-01-01T00:00:00Z"}`, "invalid order"},
		{"duplicate order", `{"spec":"squash-tree/v1","type":"squash","root":"r","base":"b","children":[{"hash":"a","order":1},{"hash":"b","order":1}],"created_at":"2026-01-01T00:00:00Z"}`, "duplicate order"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse([]byte(tt.json))
			if err == nil {
				t.Fatal("expected error")
			}
			if tt.want != "" && !strings.Contains(err.Error(), tt.want) {
				t.Errorf("error %q does not contain %q", err.Error(), tt.want)
			}
		})
	}
}

