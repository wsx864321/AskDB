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
	out := r.BuildPromptSchema("近7天订单数", "订单=orders", 1, 5000)
	if !strings.Contains(strings.ToLower(out), "create table orders") {
		t.Fatalf("expected orders in recalled schema, got %s", out)
	}
}
