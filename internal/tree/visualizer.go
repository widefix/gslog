package tree

import (
	"fmt"
	"strings"
)

type Visualizer struct {
	useColors bool
}

func NewVisualizer() *Visualizer {
	return &Visualizer{useColors: false}
}

func (v *Visualizer) Visualize(node *Node) string {
	if node == nil {
		return "(empty tree)"
	}

	var builder strings.Builder
	v.renderNode(&builder, node, "", true, true)
	return builder.String()
}

func (v *Visualizer) renderNode(builder *strings.Builder, node *Node, prefix string, isLast bool, isRoot bool) {
	var connector string
	if isRoot {
		connector = ""
	} else if isLast {
		connector = "└── "
	} else {
		connector = "├── "
	}

	var label string
	if node.IsSquash() {
		label = fmt.Sprintf("%s [SQUASH]", node.Hash)
	} else {
		label = fmt.Sprintf("%s [LEAF]", node.Hash)
	}

	builder.WriteString(prefix)
	builder.WriteString(connector)
	builder.WriteString(label)
	builder.WriteString("\n")

	var childPrefix string
	if isRoot {
		childPrefix = prefix
	} else if isLast {
		childPrefix = prefix + "    "
	} else {
		childPrefix = prefix + "│   "
	}

	for i, child := range node.Children {
		isLastChild := i == len(node.Children)-1
		v.renderNode(builder, child, childPrefix, isLastChild, false)
	}
}

func (v *Visualizer) VisualizeWithDetails(node *Node) string {
	if node == nil {
		return "(empty tree)"
	}

	var builder strings.Builder
	builder.WriteString("Squash Tree:\n")
	builder.WriteString("============\n\n")
	v.renderNodeWithDetails(&builder, node, "", true, true)
	return builder.String()
}

func (v *Visualizer) renderNodeWithDetails(builder *strings.Builder, node *Node, prefix string, isLast bool, isRoot bool) {
	var connector string
	if isRoot {
		connector = ""
	} else if isLast {
		connector = "└── "
	} else {
		connector = "├── "
	}

	var label string
	if node.IsSquash() && node.Metadata != nil {
		label = fmt.Sprintf("%s [SQUASH] base:%s strategy:%s", 
			node.Hash, 
			node.Metadata.Base, 
			node.Metadata.Strategy)
	} else {
		label = fmt.Sprintf("%s [LEAF]", node.Hash)
	}

	builder.WriteString(prefix)
	builder.WriteString(connector)
	builder.WriteString(label)
	builder.WriteString("\n")

	var childPrefix string
	if isRoot {
		childPrefix = prefix
	} else if isLast {
		childPrefix = prefix + "    "
	} else {
		childPrefix = prefix + "│   "
	}

	for i, child := range node.Children {
		isLastChild := i == len(node.Children)-1
		v.renderNodeWithDetails(builder, child, childPrefix, isLastChild, false)
	}
}
