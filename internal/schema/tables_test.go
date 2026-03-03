package schema

import "testing"

func TestExtractTableBlocks(t *testing.T) {
	s := `
CREATE TABLE users (
 id bigint,
 name varchar(10)
) ENGINE=InnoDB;

CREATE TABLE orders (
 id bigint,
 user_id bigint
) ENGINE=InnoDB;
`
	blocks := ExtractTableBlocks(s)
	if len(blocks) != 2 {
		t.Fatalf("expected 2 tables got %d", len(blocks))
	}
}
