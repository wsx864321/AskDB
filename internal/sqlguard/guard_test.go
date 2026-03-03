package sqlguard

import "testing"

func TestEnforceReadOnly(t *testing.T) {
	tests := []struct {
		name    string
		sql     string
		wantErr bool
	}{
		{name: "append limit", sql: "select * from users", wantErr: false},
		{name: "forbid update", sql: "update users set name='x'", wantErr: true},
		{name: "forbid multi", sql: "select 1; select 2", wantErr: true},
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
