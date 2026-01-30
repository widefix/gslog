package metadata

import (
	"encoding/json"
	"fmt"
)

const (
	SpecVersionV1 = "squash-tree/v1"
	TypeSquash    = "squash"
)

type ChildCommit struct {
	Hash  string `json:"hash"`
	Order int    `json:"order"`
}

type SquashMetadata struct {
	Spec     string        `json:"spec"`
	Type     string        `json:"type"`
	Root     string        `json:"root"`
	Base     string        `json:"base"`
	Children []ChildCommit `json:"children"`
	CreatedAt string       `json:"created_at"`
	Strategy string        `json:"strategy"`
}

func Parse(data []byte) (*SquashMetadata, error) {
	var metadata SquashMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, fmt.Errorf("failed to parse metadata JSON: %w", err)
	}

	if err := validate(&metadata); err != nil {
		return nil, err
	}

	return &metadata, nil
}

func validate(m *SquashMetadata) error {
	if m.Spec == "" {
		return fmt.Errorf("metadata missing required field: spec")
	}

	if m.Spec != SpecVersionV1 {
		return fmt.Errorf("unsupported spec version: %s (expected %s)", m.Spec, SpecVersionV1)
	}

	if m.Type == "" {
		return fmt.Errorf("metadata missing required field: type")
	}

	if m.Type != TypeSquash {
		return fmt.Errorf("unsupported type: %s (expected %s)", m.Type, TypeSquash)
	}

	if m.Root == "" {
		return fmt.Errorf("metadata missing required field: root")
	}

	if m.Base == "" {
		return fmt.Errorf("metadata missing required field: base")
	}

	if len(m.Children) == 0 {
		return fmt.Errorf("metadata must have at least one child commit")
	}

	seenOrders := make(map[int]bool)
	for i, child := range m.Children {
		if child.Hash == "" {
			return fmt.Errorf("child commit at index %d missing hash", i)
		}
		if child.Order < 1 {
			return fmt.Errorf("child commit at index %d has invalid order: %d (must be >= 1)", i, child.Order)
		}
		if seenOrders[child.Order] {
			return fmt.Errorf("duplicate order %d in children", child.Order)
		}
		seenOrders[child.Order] = true
	}

	return nil
}
