package retrieval

import (
	"strings"
	"testing"
)

func TestBuildPromptSchema(t *testing.T) {
	s := `
CREATE TABLE users (
 id bigint,
 name varchar(10)
) ENGINE=InnoDB;

CREATE TABLE orders (
 id bigint,
 user_id bigint,
 created_at datetime
) ENGINE=InnoDB;
`
	r := New(s)
	out := r.BuildPromptSchema("近7天订单数", "订单=orders", Options{TopK: 1, MaxBytes: 5000, KeywordBoost: 1, BM25Weight: 1, NameBoost: 8})
	if !strings.Contains(strings.ToLower(out), "create table orders") {
		t.Fatalf("expected orders in recalled schema, got %s", out)
	}
}

func TestBuildPromptSchemaFallbackDefaults(t *testing.T) {
	s := `CREATE TABLE t1 (id bigint) ENGINE=InnoDB;`
	r := New(s)
	out := r.BuildPromptSchema("", "", Options{})
	if !strings.Contains(strings.ToLower(out), "create table t1") {
		t.Fatalf("expected fallback include table, got %s", out)
	}
}
