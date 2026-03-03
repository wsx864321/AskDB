package schema

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadOptionalTextMissing(t *testing.T) {
	got, err := LoadOptionalText(filepath.Join(t.TempDir(), "none.md"), 100)
	if err != nil {
		t.Fatal(err)
	}
	if got != "" {
		t.Fatalf("expected empty for missing file, got %q", got)
	}
}

func TestLoadFewShot(t *testing.T) {
	d := t.TempDir()
	p := filepath.Join(d, "fewshot.jsonl")
	content := "{\"question\":\"q1\",\"sql\":\"select 1\"}\n# c\n{\"question\":\"q2\",\"sql\":\"select 2\"}\n"
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	examples, err := LoadFewShot(p, 1024)
	if err != nil {
		t.Fatal(err)
	}
	if len(examples) != 2 {
		t.Fatalf("expected 2 examples got %d", len(examples))
	}
}
