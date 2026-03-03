package schema

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadPromptSchema(t *testing.T) {
	d := t.TempDir()
	p := filepath.Join(d, "schema.sql")
	if err := os.WriteFile(p, []byte("CREATE TABLE t1(id int);"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := LoadPromptSchema(p, 10)
	if err != nil {
		t.Fatal(err)
	}
	if got == "" {
		t.Fatal("expected non-empty schema")
	}
}
