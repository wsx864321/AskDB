package sqlguard

import (
	"strings"
	"testing"
)

func TestEnforceReadOnly(t *testing.T) {
	tests := []struct {
		name    string
		sql     string
		wantErr bool
	}{
		{name: "append limit", sql: "select * from users", wantErr: false},
		{name: "forbid update", sql: "update users set name='x'", wantErr: true},
		{name: "forbid multi", sql: "select 1; select 2", wantErr: true},
		{name: "strip comments", sql: "-- comment\nselect * from t", wantErr: false},
		{name: "forbid into outfile", sql: "select * into outfile '/tmp/x' from users", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := EnforceReadOnly(tt.sql, 100, 1000)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err=%v wantErr=%v", err, tt.wantErr)
			}
		})
	}
}

func TestMaxLimitClamp(t *testing.T) {
	got, err := EnforceReadOnly("select * from users limit 9999", 100, 500)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(strings.ToLower(got), "limit 500") {
		t.Fatalf("expected limit clamp, got %s", got)
	}
}
