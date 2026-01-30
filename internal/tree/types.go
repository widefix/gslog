package tree

import "squash-tree/internal/metadata"

type NodeType int

const (
	NodeTypeLeaf NodeType = iota
	NodeTypeSquash
)

type Node struct {
	Hash     string
	Type     NodeType
	Metadata *metadata.SquashMetadata
	Children []*Node
	Visited  bool
}

func (n *Node) IsSquash() bool {
	return n.Type == NodeTypeSquash
}

func (n *Node) IsLeaf() bool {
	return n.Type == NodeTypeLeaf
}
