package git

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"squash-tree/internal/metadata"
)

const (
	NotesRef = "refs/notes/squash-tree"
)

type NotesReader struct {
	repoPath string
}

func NewNotesReader(repoPath string) *NotesReader {
	return &NotesReader{repoPath: repoPath}
}

func (nr *NotesReader) ReadMetadata(commitHash string) (*metadata.SquashMetadata, error) {
	shortHash, err := nr.getShortHash(commitHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get short hash: %w", err)
	}
	noteContent, err := nr.readNote(shortHash)
	if err != nil {
		return nil, fmt.Errorf("failed to read note for commit %s: %w", shortHash, err)
	}
	if noteContent == "" {
		return nil, fmt.Errorf("no squash metadata found for commit %s", shortHash)
	}
	meta, err := metadata.Parse([]byte(noteContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse metadata: %w", err)
	}
	return meta, nil
}

func (nr *NotesReader) HasMetadata(commitHash string) bool {
	shortHash, err := nr.getShortHash(commitHash)
	if err != nil {
		return false
	}

	noteContent, err := nr.readNote(shortHash)
	return err == nil && noteContent != ""
}

func (nr *NotesReader) readNote(commitHash string) (string, error) {
	cmd := exec.Command("git", "notes", "--ref", NotesRef, "show", commitHash)
	if nr.repoPath != "" {
		cmd.Dir = nr.repoPath
	}
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return "", nil
		}
		return "", fmt.Errorf("git notes show failed: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

func (nr *NotesReader) getShortHash(commitHash string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "--short", commitHash)
	if nr.repoPath != "" {
		cmd.Dir = nr.repoPath
	}

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git rev-parse failed: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

func (nr *NotesReader) CommitExists(commitHash string) bool {
	cmd := exec.Command("git", "cat-file", "-e", commitHash)
	if nr.repoPath != "" {
		cmd.Dir = nr.repoPath
	}

	err := cmd.Run()
	return err == nil
}

func WriteMetadata(repoPath, rootShortHash, baseShortHash string, children []string, strategy string) error {
	if len(children) == 0 {
		return fmt.Errorf("at least one child commit required")
	}
	childCommits := make([]metadata.ChildCommit, len(children))
	for i, h := range children {
		childCommits[i] = metadata.ChildCommit{Hash: h, Order: i + 1}
	}
	meta := &metadata.SquashMetadata{
		Spec:      metadata.SpecVersionV1,
		Type:      metadata.TypeSquash,
		Root:      rootShortHash,
		Base:      baseShortHash,
		Children:  childCommits,
		CreatedAt: time.Now().UTC().Format("2006-01-02T15:04:05Z07:00"),
		Strategy:  strategy,
	}
	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal metadata: %w", err)
	}
	cmd := exec.Command("git", "notes", "--ref", NotesRef, "add", "-F", "-", rootShortHash)
	cmd.Dir = repoPath
	cmd.Stdin = strings.NewReader(string(data))
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git notes add: %w: %s", err, string(out))
	}
	return nil
}
