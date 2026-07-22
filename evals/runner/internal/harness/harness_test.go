package harness

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// The start/poll/kill lifecycle is exercised end to end by internal/run's
// tests against the stub app; these cover manifest validation and the
// seed/read command plumbing directly.

func writeManifest(t *testing.T, dir, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, "harness.json"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestLoadErrors(t *testing.T) {
	if _, err := Load(t.TempDir()); err == nil {
		t.Error("Load without harness.json succeeded")
	}

	dir := t.TempDir()
	writeManifest(t, dir, `{"start": "x", "url": "http://localhost:1", "seed": "y"}`)
	if _, err := Load(dir); err == nil || !strings.Contains(err.Error(), `"read"`) {
		t.Errorf("Load with missing read = %v, want missing-key error", err)
	}

	writeManifest(t, dir, `not json`)
	if _, err := Load(dir); err == nil {
		t.Error("Load with bad JSON succeeded")
	}
}

func TestSeedAndReadCommands(t *testing.T) {
	dir := t.TempDir()
	writeManifest(t, dir, `{
		"start": "sleep 60",
		"url": "http://127.0.0.1:1",
		"seed": "cat > model.json",
		"read": "cat model.json"
	}`)
	app, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()

	// Seeding the empty model writes the empty array, not nothing.
	if err := app.Seed(ctx, nil); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(dir, "model.json"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(string(data)) != "[]" {
		t.Errorf("seeded empty model = %q, want []", data)
	}
	items, err := app.Read(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 0 {
		t.Errorf("Read = %v, want empty", items)
	}

	want := []Item{
		{ID: "a1", Title: "buy milk"},
		{ID: "b2", Title: "walk the dog", Completed: true},
	}
	if err := app.Seed(ctx, want); err != nil {
		t.Fatal(err)
	}
	items, err = app.Read(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(items, want) {
		t.Errorf("Read = %v, want %v", items, want)
	}
}

func TestCommandFailuresSurface(t *testing.T) {
	dir := t.TempDir()
	writeManifest(t, dir, `{
		"start": "sleep 60",
		"url": "http://127.0.0.1:1",
		"seed": "echo boom >&2; exit 3",
		"read": "echo not-json"
	}`)
	app, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := app.Seed(ctx, nil); err == nil || !strings.Contains(err.Error(), "boom") {
		t.Errorf("Seed error = %v, want failure mentioning stderr", err)
	}
	if _, err := app.Read(ctx); err == nil || !strings.Contains(err.Error(), "JSON") {
		t.Errorf("Read error = %v, want JSON parse failure", err)
	}
}
