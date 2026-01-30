package tree

import (
	"fmt"
	"sort"

	"squash-tree/internal/metadata"
)

// NotesSource provides squash metadata and commit existence for building the tree.
// *git.NotesReader implements this interface.
type NotesSource interface {
	HasMetadata(commitHash string) bool
	ReadMetadata(commitHash string) (*metadata.SquashMetadata, error)
	CommitExists(commitHash string) bool
}

type Builder struct {
	notesReader NotesSource
	visited     map[string]*Node
}

func NewBuilder(notesReader NotesSource) *Builder {
	return &Builder{
		notesReader: notesReader,
		visited:     make(map[string]*Node),
	}
}

func (b *Builder) BuildTree(commitHash string) (*Node, error) {
	b.visited = make(map[string]*Node)
	node, err := b.buildNode(commitHash)
	if err != nil {
		return nil, err
	}
	b.clearVisitedFlags(node)
	return node, nil
}

func (b *Builder) buildNode(commitHash string) (*Node, error) {
	if cached, exists := b.visited[commitHash]; exists {
		return cached, nil
	}
	if !b.notesReader.CommitExists(commitHash) {
		return nil, fmt.Errorf("commit %s does not exist", commitHash)
	}
	hasMetadata := b.notesReader.HasMetadata(commitHash)

	node := &Node{
		Hash:     commitHash,
		Children: []*Node{},
		Visited:  false,
	}
	b.visited[commitHash] = node // cache early so cycles return partial node

	if hasMetadata {
		node.Type = NodeTypeSquash
		meta, err := b.notesReader.ReadMetadata(commitHash)
		if err != nil {
			return nil, fmt.Errorf("failed to read metadata for %s: %w", commitHash, err)
		}
		node.Metadata = meta

		children := make([]metadata.ChildCommit, len(meta.Children))
		copy(children, meta.Children)
		sort.Slice(children, func(i, j int) bool {
			return children[i].Order < children[j].Order
		})

		for _, childCommit := range children {
			childNode, err := b.buildNode(childCommit.Hash)
			if err != nil {
				return nil, fmt.Errorf("failed to build child node %s: %w", childCommit.Hash, err)
			}
			if b.hasCycle(childNode, commitHash) {
				return nil, fmt.Errorf("cycle detected: commit %s is part of a cycle", childCommit.Hash)
			}
			node.Children = append(node.Children, childNode)
		}
	} else {
		node.Type = NodeTypeLeaf
	}

	return node, nil
}

func (b *Builder) hasCycle(node *Node, targetHash string) bool {
	if node.Hash == targetHash {
		return true
	}
	if node.Visited {
		return false
	}

	node.Visited = true
	defer func() { node.Visited = false }()

	for _, child := range node.Children {
		if b.hasCycle(child, targetHash) {
			return true
		}
	}

	return false
}

func (b *Builder) clearVisitedFlags(node *Node) {
	if node == nil {
		return
	}

	node.Visited = false
	for _, child := range node.Children {
		b.clearVisitedFlags(child)
	}
}
