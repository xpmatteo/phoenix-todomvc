package run

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// stubAppDir assembles a disposable app directory around the read-only stub
// server in testdata/stubapp: copies its sources, picks a free port, and
// writes the harness.json manifest the runner will drive.
func stubAppDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	for _, f := range []string{"main.go", "go.mod"} {
		data, err := os.ReadFile(filepath.Join("testdata", "stubapp", f))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, f), data, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := l.Addr().(*net.TCPAddr).Port
	l.Close()

	manifest := fmt.Sprintf(`{
  "start": "go run . -port %d",
  "url": "http://127.0.0.1:%d",
  "seed": "cat > model.json",
  "read": "cat model.json"
}
`, port, port)
	if err := os.WriteFile(filepath.Join(dir, "harness.json"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestRunEndToEndPasses(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Minute)
	defer cancel()

	var out bytes.Buffer
	results, err := Run(ctx, Options{
		AppDir:       stubAppDir(t),
		ScenariosDir: filepath.Join("testdata", "scenarios"),
		AssertWait:   2 * time.Second,
		Out:          &out,
	})
	if err != nil {
		t.Fatalf("Run: %v\noutput:\n%s", err, out.String())
	}
	if len(results) != 5 {
		t.Fatalf("got %d results, want 5\noutput:\n%s", len(results), out.String())
	}
	for _, r := range results {
		if len(r.Failures) != 0 {
			t.Errorf("scenario %q failed:\n%s", r.Scenario.Name, strings.Join(r.Failures, "\n"))
		}
	}
	if !strings.Contains(out.String(), "5 passed, 0 failed, 5 total") {
		t.Errorf("summary missing:\n%s", out.String())
	}
}

func TestRunEndToEndReportsFailures(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Minute)
	defer cancel()

	var out bytes.Buffer
	results, err := Run(ctx, Options{
		AppDir:       stubAppDir(t),
		ScenariosDir: filepath.Join("testdata", "scenarios-fail"),
		AssertWait:   500 * time.Millisecond,
		Out:          &out,
	})
	if err != nil {
		t.Fatalf("Run: %v\noutput:\n%s", err, out.String())
	}
	if len(results) != 1 || len(results[0].Failures) == 0 {
		t.Fatalf("results = %+v, want 1 failing scenario\noutput:\n%s", results, out.String())
	}

	text := out.String()
	// The page mismatch is printed as a side-by-side diff of expected and
	// actual projections, and the model mismatch is reported too.
	for _, want := range []string{
		"FAIL",
		"THEN page: projection mismatch",
		"expected",
		"actual",
		"[ ] imaginary thing",
		"[ ] real thing",
		"THEN model: mismatch",
		"1 failed",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("output missing %q:\n%s", want, text)
		}
	}
	// The diff marks differing rows.
	if !strings.Contains(text, "! ") {
		t.Errorf("no diff markers in output:\n%s", text)
	}
}
