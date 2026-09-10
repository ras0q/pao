package pao

import (
	"errors"
	"testing"
	"time"
)

func TestStateString(t *testing.T) {
	t.Parallel()

	updated := time.Date(2026, 9, 10, 15, 4, 5, 0, time.UTC)

	tests := []struct {
		name  string
		state State[int]
		want  string
	}{
		{
			name:  "empty",
			state: State[int]{},
			want:  "- empty",
		},
		{
			name:  "loading",
			state: State[int]{Fetching: true},
			want:  "* loading",
		},
		{
			name:  "error",
			state: State[int]{Err: errors.New("boom")},
			want:  "! error: boom",
		},
		{
			name:  "data",
			state: State[int]{Data: 42, HasData: true, UpdatedAt: updated},
			want:  "42 · 15:04:05",
		},
		{
			name:  "data without time",
			state: State[int]{Data: 42, HasData: true},
			want:  "42",
		},
		{
			name:  "refreshing",
			state: State[int]{Data: 42, HasData: true, Fetching: true, UpdatedAt: updated},
			want:  "42 · 15:04:05 · loading",
		},
		{
			name:  "stale error",
			state: State[int]{Data: 42, HasData: true, Err: errors.New("boom"), UpdatedAt: updated},
			want:  "42 · 15:04:05 · error: boom",
		},
		{
			name:  "stale error without time",
			state: State[int]{Data: 42, HasData: true, Err: errors.New("boom")},
			want:  "42 · error: boom",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.state.String(); got != tt.want {
				t.Fatalf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}
