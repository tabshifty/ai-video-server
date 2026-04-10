package repository

import (
	"database/sql"
	"testing"
)

func TestNullInt32ToInt(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   sql.NullInt32
		want int
	}{
		{
			name: "null value returns zero",
			in:   sql.NullInt32{},
			want: 0,
		},
		{
			name: "valid value converts to int",
			in: sql.NullInt32{
				Int32: 123,
				Valid: true,
			},
			want: 123,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := nullInt32ToInt(tt.in)
			if got != tt.want {
				t.Fatalf("nullInt32ToInt() = %d, want %d", got, tt.want)
			}
		})
	}
}
