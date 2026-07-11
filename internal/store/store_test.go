package store_test

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/AbsolutOD/zipline/internal/store"
)

func newStore(t *testing.T) *store.Store {
	t.Helper()
	st, err := store.Open(filepath.Join(t.TempDir(), "sub", "zipline.db"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	return st
}

func TestAddAndGet(t *testing.T) {
	st := newStore(t)
	a := &store.Alias{Name: "runlike", Command: "docker run --rm assaflavie/runlike", Description: "reverse-engineer docker run"}
	if err := st.Add(a); err != nil {
		t.Fatalf("Add: %v", err)
	}
	got, err := st.Get("runlike")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Command != a.Command || got.Description != a.Description || got.UseCount != 0 || got.LastUsedAt != nil {
		t.Errorf("Get returned %+v, want command=%q desc=%q usecount=0 lastused=nil", got, a.Command, a.Description)
	}
}

func TestAddDuplicateReturnsErrExists(t *testing.T) {
	st := newStore(t)
	if err := st.Add(&store.Alias{Name: "dup", Command: "docker run a"}); err != nil {
		t.Fatalf("Add: %v", err)
	}
	err := st.Add(&store.Alias{Name: "dup", Command: "docker run b"})
	if !errors.Is(err, store.ErrExists) {
		t.Errorf("Add duplicate = %v, want ErrExists", err)
	}
}

func TestGetMissingReturnsErrNotFound(t *testing.T) {
	st := newStore(t)
	_, err := st.Get("nope")
	if !errors.Is(err, store.ErrNotFound) {
		t.Errorf("Get missing = %v, want ErrNotFound", err)
	}
}

func TestListSortsByUsageThenName(t *testing.T) {
	st := newStore(t)
	for _, n := range []string{"bbb", "aaa", "ccc"} {
		if err := st.Add(&store.Alias{Name: n, Command: "docker run " + n}); err != nil {
			t.Fatalf("Add %s: %v", n, err)
		}
	}
	if err := st.Touch("ccc"); err != nil {
		t.Fatalf("Touch: %v", err)
	}
	got, err := st.List("usage")
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(got) != 3 || got[0].Name != "ccc" || got[1].Name != "aaa" || got[2].Name != "bbb" {
		t.Errorf("List(usage) order = %v, want [ccc aaa bbb]", names(got))
	}
	got, err = st.List("name")
	if err != nil {
		t.Fatalf("List(name): %v", err)
	}
	if len(got) != 3 || got[0].Name != "aaa" || got[1].Name != "bbb" || got[2].Name != "ccc" {
		t.Errorf("List(name) order = %v, want [aaa bbb ccc]", names(got))
	}
}

func names(as []store.Alias) []string {
	out := make([]string, len(as))
	for i, a := range as {
		out[i] = a.Name
	}
	return out
}

func TestUpdate(t *testing.T) {
	st := newStore(t)
	if err := st.Add(&store.Alias{Name: "edit-me", Command: "docker run old"}); err != nil {
		t.Fatalf("Add: %v", err)
	}
	a, err := st.Get("edit-me")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	a.Command = "docker run new"
	if err := st.Update(a); err != nil {
		t.Fatalf("Update: %v", err)
	}
	got, err := st.Get("edit-me")
	if err != nil {
		t.Fatalf("Get after update: %v", err)
	}
	if got.Command != "docker run new" {
		t.Errorf("Command after update = %q, want %q", got.Command, "docker run new")
	}
}

func TestDelete(t *testing.T) {
	st := newStore(t)
	if err := st.Add(&store.Alias{Name: "bye", Command: "docker run x"}); err != nil {
		t.Fatalf("Add: %v", err)
	}
	if err := st.Delete("bye"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := st.Get("bye"); !errors.Is(err, store.ErrNotFound) {
		t.Errorf("Get after delete = %v, want ErrNotFound", err)
	}
	if err := st.Delete("bye"); !errors.Is(err, store.ErrNotFound) {
		t.Errorf("Delete missing = %v, want ErrNotFound", err)
	}
}

func TestTouch(t *testing.T) {
	st := newStore(t)
	if err := st.Add(&store.Alias{Name: "used", Command: "docker run x"}); err != nil {
		t.Fatalf("Add: %v", err)
	}
	if err := st.Touch("used"); err != nil {
		t.Fatalf("Touch: %v", err)
	}
	if err := st.Touch("used"); err != nil {
		t.Fatalf("Touch 2: %v", err)
	}
	got, err := st.Get("used")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.UseCount != 2 {
		t.Errorf("UseCount = %d, want 2", got.UseCount)
	}
	if got.LastUsedAt == nil {
		t.Error("LastUsedAt = nil, want set")
	}
	if err := st.Touch("missing"); !errors.Is(err, store.ErrNotFound) {
		t.Errorf("Touch missing = %v, want ErrNotFound", err)
	}
}

func TestOpenIsIdempotent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "zipline.db")
	st, err := store.Open(path)
	if err != nil {
		t.Fatalf("first Open: %v", err)
	}
	if err := st.Add(&store.Alias{Name: "keep", Command: "docker run x"}); err != nil {
		t.Fatalf("Add: %v", err)
	}
	st2, err := store.Open(path)
	if err != nil {
		t.Fatalf("second Open: %v", err)
	}
	if _, err := st2.Get("keep"); err != nil {
		t.Errorf("Get after reopen: %v", err)
	}
}
