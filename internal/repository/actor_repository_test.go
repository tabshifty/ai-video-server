package repository

import "testing"

func TestNormalizeActorName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "trim and lowercase",
			in:   "  Alice  Bob  ",
			want: "alice bob",
		},
		{
			name: "collapse spaces",
			in:   "张三   李四",
			want: "张三 李四",
		},
		{
			name: "empty input",
			in:   "   ",
			want: "",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := normalizeActorName(tt.in)
			if got != tt.want {
				t.Fatalf("normalizeActorName() = %q, want %q", got, tt.want)
			}
		})
	}
}
