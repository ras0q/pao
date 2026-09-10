package pao

import (
	"fmt"
	"time"
)

// State is the current resource data and fetch status.
type State[T any] struct {
	Data        T
	HasData     bool
	Fetching    bool
	Invalidated bool
	Err         error
	UpdatedAt   time.Time
}

// String returns a human-readable summary for display in View.
func (s State[T]) String() string {
	updatedAt := ""
	if !s.UpdatedAt.IsZero() {
		updatedAt = " · " + s.UpdatedAt.Format("15:04:05")
	}

	switch {
	case s.HasData && s.Err != nil:
		return fmt.Sprintf("%v%s · error: %s", s.Data, updatedAt, s.Err.Error())
	case s.Err != nil:
		return fmt.Sprintf("! error: %s", s.Err.Error())
	case s.HasData && s.Fetching:
		return fmt.Sprintf("%v%s · loading", s.Data, updatedAt)
	case s.HasData:
		return fmt.Sprintf("%v%s", s.Data, updatedAt)
	case s.Fetching:
		return "* loading"
	default:
		return "- empty"
	}
}
