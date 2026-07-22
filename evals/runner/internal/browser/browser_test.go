package browser

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"todomvc/evals/runner/internal/scenario"
)

// fixtureServer serves the static fixture pages. The root and filter routes
// serve list variants so filter clicks, `go to`, and `reload` traverse real
// full-page navigations like a server-rendered app would.
func fixtureServer(t *testing.T) *httptest.Server {
	t.Helper()
	dir := "testdata"
	page := func(name string) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, filepath.Join(dir, name))
		}
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/app.css", page("app.css"))
	mux.HandleFunc("/{page}", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(dir, r.PathValue("page")))
	})
	mux.HandleFunc("/{$}", page("list.html"))
	mux.HandleFunc("/active", page("list-active.html"))
	mux.HandleFunc("/completed", page("list-completed.html"))
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func newTestSession(t *testing.T) *Session {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	t.Cleanup(cancel)
	s, err := NewSession(ctx, "")
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}
	t.Cleanup(s.Close)
	return s
}

func navigate(t *testing.T, s *Session, url string) {
	t.Helper()
	if err := s.Navigate(url); err != nil {
		t.Fatalf("Navigate(%s): %v", url, err)
	}
}

func project(t *testing.T, s *Session) []string {
	t.Helper()
	lines, err := s.Project()
	if err != nil {
		t.Fatalf("Project: %v", err)
	}
	return lines
}

func do(t *testing.T, s *Session, base string, raw string) {
	t.Helper()
	a, err := scenario.ParseAction(raw)
	if err != nil {
		t.Fatalf("ParseAction(%q): %v", raw, err)
	}
	if err := s.Do(a, base); err != nil {
		t.Fatalf("Do(%q): %v", raw, err)
	}
}

func check(t *testing.T, s *Session, raw string) (bool, string) {
	t.Helper()
	c, err := scenario.ParseCheck(raw)
	if err != nil {
		t.Fatalf("ParseCheck(%q): %v", raw, err)
	}
	ok, detail, err := s.Check(c)
	if err != nil {
		t.Fatalf("Check(%q): %v", raw, err)
	}
	return ok, detail
}

// checkEventually polls a check until it reports the wanted value, the same
// way the runner polls THEN assertions (autofocus, for one, lands a beat
// after navigation).
func checkEventually(t *testing.T, s *Session, raw string, want bool) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for {
		ok, detail := check(t, s, raw)
		if ok == want {
			return
		}
		if time.Now().After(deadline) {
			t.Errorf("check %q = %v (%s), want %v", raw, ok, detail, want)
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
}

func TestProjectionFixtures(t *testing.T) {
	srv := fixtureServer(t)
	s := newTestSession(t)

	cases := []struct {
		path string
		want []string
	}{
		{"/empty.html", []string{">"}},
		{"/", []string{
			"v >",
			"#a1 [x] ~buy milk~",
			"#b2 [ ] walk the dog",
			"-- **1** item left | (All) Active Completed | [Clear completed]",
		}},
		{"/active", []string{
			"v >",
			"#b2 [ ] walk the dog",
			"-- **1** item left | All (Active) Completed | [Clear completed]",
		}},
		{"/completed", []string{
			"v >",
			"#a1 [x] ~buy milk~",
			"-- **1** item left | All Active (Completed) | [Clear completed]",
		}},
		{"/editing.html", []string{
			"v >",
			"#a1 [edit: buy oat milk]",
			"#b2 [ ] walk the dog",
			"-- **2** items left | (All) Active Completed",
		}},
		{"/allcomplete.html", []string{
			"(v) > buy mil",
			"#a1 [x] ~buy milk~",
			"-- **0** items left | (All) Active Completed | [Clear completed]",
		}},
		// A buggy app: [x] and ~…~ are independent signals; an editing row
		// whose view is still shown yields both lines; a row without a
		// data-id renders without the # prefix; a footer without filter
		// links or a Clear-completed button is just the counter.
		{"/buggy.html", []string{
			"v >",
			"#c1 [x] checked not styled",
			"#c2 [ ] ~styled not checked~",
			"[ ] no id here",
			"#c4 [ ] half editing",
			"#c4 [edit: half editing now]",
			"-- **3** items left",
		}},
	}
	for _, c := range cases {
		t.Run(c.path, func(t *testing.T) {
			navigate(t, s, srv.URL+c.path)
			if got := project(t, s); !reflect.DeepEqual(got, c.want) {
				t.Errorf("projection of %s =\n%q\nwant\n%q", c.path, got, c.want)
			}
		})
	}
}

func TestDisplayedItems(t *testing.T) {
	srv := fixtureServer(t)
	s := newTestSession(t)

	navigate(t, s, srv.URL+"/")
	got, err := s.DisplayedItems()
	if err != nil {
		t.Fatal(err)
	}
	want := []DisplayedItem{
		{Title: "buy milk", ID: "a1", HasID: true},
		{Title: "walk the dog", ID: "b2", HasID: true},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("DisplayedItems = %v, want %v", got, want)
	}

	navigate(t, s, srv.URL+"/buggy.html")
	got, err = s.DisplayedItems()
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 4 || got[2].HasID || got[2].Title != "no id here" {
		t.Errorf("buggy DisplayedItems = %v, want 4 items with [2] lacking an id", got)
	}
}

func TestTypeAndClear(t *testing.T) {
	srv := fixtureServer(t)
	s := newTestSession(t)
	navigate(t, s, srv.URL+"/empty.html")

	// The fixture input has autofocus, so typing goes into it once focus
	// has landed.
	checkEventually(t, s, "focus is on the new-todo input", true)
	do(t, s, srv.URL, `type "buy milk"`)
	if got := project(t, s); !reflect.DeepEqual(got, []string{"> buy milk"}) {
		t.Errorf("after type: projection = %q", got)
	}
	do(t, s, srv.URL, "clear")
	if got := project(t, s); !reflect.DeepEqual(got, []string{">"}) {
		t.Errorf("after clear: projection = %q", got)
	}
	// Whitespace-heavy text survives typing verbatim.
	do(t, s, srv.URL, `type "   buy milk   "`)
	if got := project(t, s); !reflect.DeepEqual(got, []string{">    buy milk"}) {
		t.Errorf("after padded type: projection = %q", got)
	}
}

func TestFocusChecksAndBlur(t *testing.T) {
	srv := fixtureServer(t)
	s := newTestSession(t)

	navigate(t, s, srv.URL+"/empty.html")
	checkEventually(t, s, "focus is on the new-todo input", true)
	do(t, s, srv.URL, "blur")
	checkEventually(t, s, "focus is on the new-todo input", false)

	// The editing fixture autofocuses the edit field.
	navigate(t, s, srv.URL+"/editing.html")
	checkEventually(t, s, `focus is in the edit field of "buy milk"`, true)
	checkEventually(t, s, `focus is in the edit field of "walk the dog"`, false)
}

func TestHoverAndDestroyVisibility(t *testing.T) {
	srv := fixtureServer(t)
	s := newTestSession(t)
	navigate(t, s, srv.URL+"/")

	// Before any hover, both destroy buttons are hidden (CSS default).
	if ok, detail := check(t, s, `destroy button of "buy milk" is hidden`); !ok {
		t.Errorf("pre-hover: %s", detail)
	}
	do(t, s, srv.URL, `hover "buy milk"`)
	if ok, detail := check(t, s, `destroy button of "buy milk" is visible`); !ok {
		t.Errorf("post-hover: %s", detail)
	}
	if ok, detail := check(t, s, `destroy button of "walk the dog" is hidden`); !ok {
		t.Errorf("post-hover, other row: %s", detail)
	}
}

func TestClickToggleAndToggleAll(t *testing.T) {
	srv := fixtureServer(t)
	s := newTestSession(t)
	navigate(t, s, srv.URL+"/")

	// A native checkbox toggles even on a static page. The completed
	// styling does not follow (no app JS), which the projection reports
	// honestly: [x] without ~…~.
	do(t, s, srv.URL, `click toggle of "walk the dog"`)
	got := project(t, s)
	if got[2] != "#b2 [x] walk the dog" {
		t.Errorf("after toggle click: line = %q, want %q", got[2], "#b2 [x] walk the dog")
	}

	do(t, s, srv.URL, "click toggle-all")
	got = project(t, s)
	if got[0] != "(v) >" {
		t.Errorf("after toggle-all click: input row = %q, want %q", got[0], "(v) >")
	}
}

func TestFilterClickNavigatesAndReload(t *testing.T) {
	srv := fixtureServer(t)
	s := newTestSession(t)
	navigate(t, s, srv.URL+"/")

	do(t, s, srv.URL, `click filter "Active"`)
	want := []string{
		"v >",
		"#b2 [ ] walk the dog",
		"-- **1** item left | All (Active) Completed | [Clear completed]",
	}
	if got := project(t, s); !reflect.DeepEqual(got, want) {
		t.Errorf("after filter click: projection = %q, want %q", got, want)
	}

	do(t, s, srv.URL, "reload")
	if got := project(t, s); !reflect.DeepEqual(got, want) {
		t.Errorf("after reload: projection = %q, want %q", got, want)
	}

	do(t, s, srv.URL, `go to "/completed"`)
	want = []string{
		"v >",
		"#a1 [x] ~buy milk~",
		"-- **1** item left | All Active (Completed) | [Clear completed]",
	}
	if got := project(t, s); !reflect.DeepEqual(got, want) {
		t.Errorf("after go to: projection = %q, want %q", got, want)
	}
}

// TestInertActionsDispatch drives the verbs whose effects need app-side
// behavior a static fixture cannot supply (dblclick entering edit mode,
// Enter/Escape handling, destroy removing a row, Clear completed). Here we
// can only assert that dispatching them against real matching elements
// succeeds; their semantics are exercised end-to-end in internal/run tests
// and, ultimately, only against a real implementation.
func TestInertActionsDispatch(t *testing.T) {
	srv := fixtureServer(t)
	s := newTestSession(t)
	navigate(t, s, srv.URL+"/")

	for _, raw := range []string{
		`dblclick "buy milk"`,
		`click destroy of "buy milk"`,
		`click "Clear completed"`,
		"press Enter",
		"press Escape",
	} {
		a, err := scenario.ParseAction(raw)
		if err != nil {
			t.Fatalf("ParseAction(%q): %v", raw, err)
		}
		if err := s.Do(a, srv.URL); err != nil {
			t.Errorf("Do(%q): %v", raw, err)
		}
	}
}

func TestActionsFailOnMissingTargets(t *testing.T) {
	srv := fixtureServer(t)
	s := newTestSession(t)
	navigate(t, s, srv.URL+"/")

	for _, raw := range []string{
		`click toggle of "no such todo"`,
		`hover "nope"`,
		`click filter "Bogus"`, // parseable? no — filter names are closed; use a missing todo instead
	} {
		a, err := scenario.ParseAction(raw)
		if err != nil {
			continue // closed-vocabulary rejection is also a correct outcome
		}
		if err := s.Do(a, srv.URL); err == nil {
			t.Errorf("Do(%q) succeeded, want error", raw)
		}
	}
}
