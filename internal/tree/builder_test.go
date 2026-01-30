package tree

import (
	"strings"
	"testing"

	"squash-tree/internal/metadata"
)

// mockNotesSource is a test double for NotesSource.
type mockNotesSource struct {
	commits   map[string]bool
	metadata  map[string]*metadata.SquashMetadata
	hasMeta   map[string]bool
}

func newMockNotesSource() *mockNotesSource {
	return &mockNotesSource{
		commits:  make(map[string]bool),
		metadata: make(map[string]*metadata.SquashMetadata),
		hasMeta:  make(map[string]bool),
	}
}

func (m *mockNotesSource) addCommit(hash string) {
	m.commits[hash] = true
}

func (m *mockNotesSource) addSquash(root, base string, childHashes []string) {
	m.commits[root] = true
	m.hasMeta[root] = true
	children := make([]metadata.ChildCommit, len(childHashes))
	for i, h := range childHashes {
		children[i] = metadata.ChildCommit{Hash: h, Order: i + 1}
	}
	m.metadata[root] = &metadata.SquashMetadata{
		Spec:     metadata.SpecVersionV1,
		Type:     metadata.TypeSquash,
		Root:     root,
		Base:     base,
		Children: children,
	}
}

func (m *mockNotesSource) CommitExists(commitHash string) bool {
	return m.commits[commitHash]
}

func (m *mockNotesSource) HasMetadata(commitHash string) bool {
	return m.hasMeta[commitHash]
}

func (m *mockNotesSource) ReadMetadata(commitHash string) (*metadata.SquashMetadata, error) {
	meta, ok := m.metadata[commitHash]
	if !ok {
		return nil, nil
	}
	return meta, nil
}

func TestBuilder_LeafOnly(t *testing.T) {
	mock := newMockNotesSource()
	mock.addCommit("abc")
	b := NewBuilder(mock)

	node, err := b.BuildTree("abc")
	if err != nil {
		t.Fatalf("BuildTree: %v", err)
	}
	if node.Hash != "abc" {
		t.Errorf("Hash: got %q", node.Hash)
	}
	if node.Type != NodeTypeLeaf {
		t.Errorf("Type: got %v, want Leaf", node.Type)
	}
	if len(node.Children) != 0 {
		t.Errorf("Children: got %d", len(node.Children))
	}
}

func TestBuilder_SingleSquashWithChildren(t *testing.T) {
	mock := newMockNotesSource()
	mock.addCommit("base")
	mock.addCommit("c1")
	mock.addCommit("c2")
	mock.addSquash("root", "base", []string{"c1", "c2"})
	b := NewBuilder(mock)

	node, err := b.BuildTree("root")
	if err != nil {
		t.Fatalf("BuildTree: %v", err)
	}
	if node.Type != NodeTypeSquash {
		t.Fatalf("Type: got %v", node.Type)
	}
	if len(node.Children) != 2 {
		t.Fatalf("Children: got %d", len(node.Children))
	}
	if node.Children[0].Hash != "c1" || node.Children[1].Hash != "c2" {
		t.Errorf("Children hashes: %q %q", node.Children[0].Hash, node.Children[1].Hash)
	}
	if !node.Children[0].IsLeaf() || !node.Children[1].IsLeaf() {
		t.Error("children should be leaves")
	}
}

func TestBuilder_NestedSquash(t *testing.T) {
	mock := newMockNotesSource()
	mock.addCommit("base")
	mock.addCommit("leaf1")
	mock.addSquash("inner", "base", []string{"leaf1"})
	mock.addSquash("root", "base", []string{"inner"})
	b := NewBuilder(mock)

	node, err := b.BuildTree("root")
	if err != nil {
		t.Fatalf("BuildTree: %v", err)
	}
	if node.Type != NodeTypeSquash || len(node.Children) != 1 {
		t.Fatalf("root: Type=%v Children=%d", node.Type, len(node.Children))
	}
	inner := node.Children[0]
	if inner.Hash != "inner" || inner.Type != NodeTypeSquash {
		t.Fatalf("inner: Hash=%q Type=%v", inner.Hash, inner.Type)
	}
	if len(inner.Children) != 1 || inner.Children[0].Hash != "leaf1" {
		t.Fatalf("inner.Children: %+v", inner.Children)
	}
}

func TestBuilder_MissingCommit(t *testing.T) {
	mock := newMockNotesSource()
	mock.addCommit("exists")
	b := NewBuilder(mock)

	_, err := b.BuildTree("nonexistent")
	if err == nil {
		t.Fatal("BuildTree: expected error for missing commit")
	}
	if !strings.Contains(err.Error(), "does not exist") {
		t.Errorf("error %q", err.Error())
	}
}

func TestBuilder_CycleDetection(t *testing.T) {
	mock := newMockNotesSource()
	mock.addCommit("base")
	// root -> mid -> root (cycle)
	mock.addSquash("root", "base", []string{"mid"})
	mock.addSquash("mid", "base", []string{"root"})
	b := NewBuilder(mock)

	_, err := b.BuildTree("root")
	if err == nil {
		t.Fatal("BuildTree: expected error for cycle")
	}
	if !strings.Contains(err.Error(), "cycle") {
		t.Errorf("error %q", err.Error())
	}
}
