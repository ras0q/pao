package pao

import (
	"errors"
	"reflect"
	"testing"

	tea "charm.land/bubbletea/v2"
)

func run(cmd tea.Cmd) tea.Msg {
	if cmd == nil {
		return nil
	}
	return cmd()
}

func mustLoaded[T any](t *testing.T, msg tea.Msg) LoadedMsg[T] {
	t.Helper()

	loaded, ok := msg.(LoadedMsg[T])
	if !ok {
		t.Fatalf("expected LoadedMsg[%T], got %T", *new(T), msg)
	}
	return loaded
}

func assertNoMsg(t *testing.T, msg tea.Msg) {
	t.Helper()
	if msg != nil {
		t.Fatalf("expected nil msg, got %T", msg)
	}
}

func assertState[T any](t *testing.T, got, want State[T]) {
	t.Helper()

	if got.HasData != want.HasData {
		t.Fatalf("HasData = %v, want %v (full: %+v)", got.HasData, want.HasData, got)
	}
	if got.Fetching != want.Fetching {
		t.Fatalf("Fetching = %v, want %v (full: %+v)", got.Fetching, want.Fetching, got)
	}
	if got.Invalidated != want.Invalidated {
		t.Fatalf("Invalidated = %v, want %v (full: %+v)", got.Invalidated, want.Invalidated, got)
	}
	if want.HasData && !reflect.DeepEqual(got.Data, want.Data) {
		t.Fatalf("Data = %v, want %v (full: %+v)", got.Data, want.Data, got)
	}

	switch {
	case want.Err != nil:
		if !errors.Is(got.Err, want.Err) {
			t.Fatalf("Err = %v, want %v (full: %+v)", got.Err, want.Err, got)
		}
	case got.Err != nil:
		t.Fatalf("Err = %v, want nil (full: %+v)", got.Err, got)
	}
}
