package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	chunker "github.com/voznikaetnepriyazn/Hack-Change/agent/rag/chunk"
)

type recordingPipeline struct {
	documents []chunker.Document
	err       error
}

func (p *recordingPipeline) IngestDocument(_ context.Context, doc chunker.Document) error {
	if p.err != nil {
		return p.err
	}
	p.documents = append(p.documents, doc)
	return nil
}

func TestIngestMarkdownDirectoryLoadsEachDocumentWithSourceMetadata(t *testing.T) {
	root := t.TempDir()
	files := map[string]string{
		"root.md":                  "root document",
		"playbook-master/guide.md": "playbook document",
		"synthetic/policy.MD":      "synthetic document",
		"synthetic/ignored.txt":    "not a Markdown document",
		"playbook-master/.gitkeep": "",
	}
	for name, content := range files {
		path := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			t.Fatalf("create document directory: %v", err)
		}
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatalf("write test document: %v", err)
		}
	}

	pipeline := &recordingPipeline{}
	count, err := ingestMarkdownDirectory(context.Background(), root, pipeline)
	if err != nil {
		t.Fatalf("ingestMarkdownDirectory() returned error: %v", err)
	}
	if count != 3 || len(pipeline.documents) != 3 {
		t.Fatalf("ingested documents = %d, want 3", count)
	}

	gotBySource := make(map[string]chunker.Document, len(pipeline.documents))
	for _, doc := range pipeline.documents {
		source, ok := doc.Metadata["source"].(string)
		if !ok {
			t.Fatalf("document metadata source = %v, want string", doc.Metadata["source"])
		}
		gotBySource[source] = doc
	}
	for source, content := range map[string]string{
		"root.md":                  "root document",
		"playbook-master/guide.md": "playbook document",
		"synthetic/policy.MD":      "synthetic document",
	} {
		doc, ok := gotBySource[source]
		if !ok {
			t.Errorf("missing ingested source %q", source)
			continue
		}
		if doc.Content != content {
			t.Errorf("document %q content = %q, want %q", source, doc.Content, content)
		}
	}
	if got := gotBySource["playbook-master/guide.md"].Metadata["category"]; got != "playbook-master" {
		t.Errorf("playbook category = %v, want playbook-master", got)
	}
	if got := gotBySource["root.md"].Metadata["category"]; got != nil {
		t.Errorf("root document category = %v, want no category", got)
	}
}

func TestIngestMarkdownDirectoryReturnsPipelineErrors(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "guide.md"), []byte("document"), 0o600); err != nil {
		t.Fatalf("write test document: %v", err)
	}
	wantErr := errors.New("embedding failed")

	_, err := ingestMarkdownDirectory(context.Background(), root, &recordingPipeline{err: wantErr})
	if !errors.Is(err, wantErr) {
		t.Errorf("ingestMarkdownDirectory() error = %v, want %v", err, wantErr)
	}
}

func TestIngestMarkdownDirectoryRequiresMarkdownFiles(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "notes.txt"), []byte("not Markdown"), 0o600); err != nil {
		t.Fatalf("write test file: %v", err)
	}

	_, err := ingestMarkdownDirectory(context.Background(), root, &recordingPipeline{})
	if err == nil || !strings.Contains(err.Error(), "no Markdown documents found") {
		t.Errorf("ingestMarkdownDirectory() error = %v, want no-Markdown-documents error", err)
	}
}
