package tree

import (
	"squash-tree/internal/metadata"
	"strings"
	"testing"
)

func TestVisualize_EmptyTree(t *testing.T) {
	v := NewVisualizer()
	out := v.Visualize(nil)
	if out != "(empty tree)" {
		t.Errorf("Visualize(nil): got %q", out)
	}
}

func TestVisualize_SingleLeaf(t *testing.T) {
	node := &Node{Hash: "abc123", Type: NodeTypeLeaf, Children: nil}
	v := NewVisualizer()
	out := v.Visualize(node)
	if !strings.Contains(out, "abc123") {
		t.Errorf("output missing hash: %q", out)
	}
	if !strings.Contains(out, "[LEAF]") {
		t.Errorf("output missing LEAF: %q", out)
	}
}

func TestVisualize_SquashWithChildren(t *testing.T) {
	root := &Node{
		Hash:     "root",
		Type:     NodeTypeSquash,
		Metadata: &metadata.SquashMetadata{Root: "root"},
		Children: []*Node{
			{Hash: "c1", Type: NodeTypeLeaf, Children: nil},
			{Hash: "c2", Type: NodeTypeLeaf, Children: nil},
		},
	}
	v := NewVisualizer()
	out := v.Visualize(root)
	if !strings.Contains(out, "root") || !strings.Contains(out, "[SQUASH]") {
		t.Errorf("output missing root/SQUASH: %q", out)
	}
	if !strings.Contains(out, "c1") || !strings.Contains(out, "c2") {
		t.Errorf("output missing children: %q", out)
	}
	if !strings.Contains(out, "[LEAF]") {
		t.Errorf("output missing LEAF: %q", out)
	}
	// Tree connectors
	if !strings.Contains(out, "├──") && !strings.Contains(out, "└──") {
		t.Errorf("output missing tree connectors: %q", out)
	}
}

func TestVisualize_NestedTree(t *testing.T) {
	inner := &Node{
		Hash:     "inner",
		Type:     NodeTypeSquash,
		Children: []*Node{{Hash: "leaf", Type: NodeTypeLeaf, Children: nil}},
	}
	root := &Node{
		Hash:     "root",
		Type:     NodeTypeSquash,
		Children: []*Node{inner},
	}
	v := NewVisualizer()
	out := v.Visualize(root)
	if !strings.Contains(out, "root") || !strings.Contains(out, "inner") || !strings.Contains(out, "leaf") {
		t.Errorf("output missing nodes: %q", out)
	}
}
