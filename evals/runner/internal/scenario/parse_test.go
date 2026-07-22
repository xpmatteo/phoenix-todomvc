package scenario

import (
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// realScenariosDir points at the durable scenario suite this runner exists for.
const realScenariosDir = "../../../scenarios"

func parseReal(t *testing.T) []Scenario {
	t.Helper()
	scs, err := ParseDir(realScenariosDir)
	if err != nil {
		t.Fatalf("ParseDir(%s): %v", realScenariosDir, err)
	}
	return scs
}

func TestParseRealSuiteCounts(t *testing.T) {
	scs := parseReal(t)
	perFile := map[string]int{}
	for _, s := range scs {
		perFile[filepath.Base(s.File)]++
	}
	want := map[string]int{
		"clear-completed.md": 2,
		"counter.md":         5,
		"editing.md":         7,
		"item.md":            4,
		"mark-all.md":        5,
		"new-todo.md":        4,
		"persistence.md":     3,
		"routing.md":         5,
	}
	if !reflect.DeepEqual(perFile, want) {
		t.Errorf("scenario counts per file = %v, want %v", perFile, want)
	}
	if len(scs) != 35 {
		t.Errorf("total scenarios = %d, want 35", len(scs))
	}
	for _, s := range scs {
		if !s.HasThenPage && !s.HasThenModel && len(s.Checks) == 0 {
			t.Errorf("%s %q: no THEN section", s.File, s.Name)
		}
	}
}

func findScenario(t *testing.T, scs []Scenario, name string) Scenario {
	t.Helper()
	for _, s := range scs {
		if s.Name == name {
			return s
		}
	}
	t.Fatalf("scenario %q not found", name)
	return Scenario{}
}

func TestParseNewTodoStructure(t *testing.T) {
	scs := parseReal(t)

	s := findScenario(t, scs, "The input is focused on load, and an empty app hides main and footer")
	if s.Given == nil || len(s.Given) != 0 {
		t.Errorf("Given = %v, want empty model", s.Given)
	}
	if s.Route != "/" {
		t.Errorf("Route = %q, want /", s.Route)
	}
	if len(s.When) != 0 {
		t.Errorf("When = %v, want none", s.When)
	}
	if !s.HasThenPage || !reflect.DeepEqual(s.ThenPage, []string{">"}) {
		t.Errorf("ThenPage = %q, want [\">\"]", s.ThenPage)
	}
	if s.HasThenModel {
		t.Error("HasThenModel = true, want false")
	}
	wantChecks := []Check{{Kind: CheckFocusNewTodo, Raw: "focus is on the new-todo input"}}
	if !reflect.DeepEqual(s.Checks, wantChecks) {
		t.Errorf("Checks = %v, want %v", s.Checks, wantChecks)
	}

	s = findScenario(t, scs, "Pressing Enter creates the todo, appends it to the list, and clears the input")
	wantWhen := []Action{
		{Verb: VerbType, Arg: "walk the dog", Raw: `type "walk the dog"`},
		{Verb: VerbPressEnter, Raw: "press Enter"},
	}
	if !reflect.DeepEqual(s.When, wantWhen) {
		t.Errorf("When = %v, want %v", s.When, wantWhen)
	}
	wantPage := []string{
		"v >",
		"[ ] buy milk",
		"[ ] walk the dog",
		"-- **2** items left | (All) Active Completed",
	}
	if !reflect.DeepEqual(s.ThenPage, wantPage) {
		t.Errorf("ThenPage = %q, want %q", s.ThenPage, wantPage)
	}
	wantModel := Model{
		{Title: "buy milk"},
		{Title: "walk the dog"},
	}
	if !s.HasThenModel || !reflect.DeepEqual(s.ThenModel, wantModel) {
		t.Errorf("ThenModel = %v, want %v", s.ThenModel, wantModel)
	}

	// Whitespace inside quoted text survives parsing.
	s = findScenario(t, scs, "The title is trimmed before the todo is created")
	if s.When[0].Arg != "   buy milk   " {
		t.Errorf("type arg = %q, want padded text", s.When[0].Arg)
	}
}

func TestParseModelIDs(t *testing.T) {
	scs := parseReal(t)

	s := findScenario(t, scs, "The destroy button removes its todo")
	wantGiven := Model{
		{ID: "a1", Title: "buy milk"},
		{ID: "b2", Title: "walk the dog"},
	}
	if !reflect.DeepEqual(s.Given, wantGiven) {
		t.Errorf("Given = %v, want %v", s.Given, wantGiven)
	}
	wantModel := Model{{ID: "b2", Title: "walk the dog"}}
	if !reflect.DeepEqual(s.ThenModel, wantModel) {
		t.Errorf("ThenModel = %v, want %v", s.ThenModel, wantModel)
	}
	// The expected page carries an id on the item line.
	if got := s.ThenPage[1]; got != "#b2 [ ] walk the dog" {
		t.Errorf("ThenPage[1] = %q", got)
	}

	s = findScenario(t, scs, "Persisted todos survive a reload, states and identities intact")
	wantGiven = Model{
		{ID: "a1", Title: "buy milk"},
		{ID: "b2", Title: "walk the dog", Completed: true},
	}
	if !reflect.DeepEqual(s.Given, wantGiven) {
		t.Errorf("Given = %v, want %v", s.Given, wantGiven)
	}
	// Mixed: ids in GIVEN and THEN model, none on the THEN page lines.
	for _, l := range s.ThenPage {
		if strings.HasPrefix(l, "#") {
			t.Errorf("unexpected id on page line %q", l)
		}
	}
}

func TestParseRoutes(t *testing.T) {
	scs := parseReal(t)
	s := findScenario(t, scs, "The Active filter shows only active todos")
	if s.Route != "/active" {
		t.Errorf("Route = %q, want /active", s.Route)
	}
	s = findScenario(t, scs, "The Completed filter shows only completed todos")
	if s.Route != "/completed" {
		t.Errorf("Route = %q, want /completed", s.Route)
	}
	s = findScenario(t, scs, "Clicking a filter link filters the list and moves the selection")
	if s.Route != "/" {
		t.Errorf("Route = %q, want default /", s.Route)
	}
	if want := []Action{{Verb: VerbFilter, Arg: "Active", Raw: `click filter "Active"`}}; !reflect.DeepEqual(s.When, want) {
		t.Errorf("When = %v, want %v", s.When, want)
	}
}

func TestParseActionVocabularyCoverage(t *testing.T) {
	scs := parseReal(t)
	used := map[Verb]bool{}
	for _, s := range scs {
		for _, a := range s.When {
			used[a.Verb] = true
		}
	}
	// Verbs the real suite exercises today; parsing must have produced them.
	for _, v := range []Verb{
		VerbType, VerbPressEnter, VerbPressEscape, VerbClear, VerbBlur,
		VerbToggle, VerbDestroy, VerbDblclick, VerbToggleAll,
		VerbClearCompleted, VerbFilter, VerbReload, VerbHover,
	} {
		if !used[v] {
			t.Errorf("verb %s not found anywhere in the real suite", v)
		}
	}
}

func TestParseChecksInRealSuite(t *testing.T) {
	scs := parseReal(t)
	s := findScenario(t, scs, "The destroy button appears on hover")
	want := []Check{
		{Kind: CheckDestroyVisible, Title: "buy milk", Raw: `destroy button of "buy milk" is visible`},
		{Kind: CheckDestroyHidden, Title: "walk the dog", Raw: `destroy button of "walk the dog" is hidden`},
	}
	if !reflect.DeepEqual(s.Checks, want) {
		t.Errorf("Checks = %v, want %v", s.Checks, want)
	}
	s = findScenario(t, scs, "Double-clicking a label enters editing mode")
	want = []Check{{Kind: CheckFocusEditField, Title: "buy milk", Raw: `focus is in the edit field of "buy milk"`}}
	if !reflect.DeepEqual(s.Checks, want) {
		t.Errorf("Checks = %v, want %v", s.Checks, want)
	}
}

func parseString(t *testing.T, doc string) ([]Scenario, error) {
	t.Helper()
	return Parse("test.md", strings.NewReader(doc))
}

func TestParseErrors(t *testing.T) {
	cases := []struct {
		name, doc, wantErr string
	}{
		{"missing GIVEN model", "## s\n\nTHEN page:\n\n    >\n", "missing required GIVEN model"},
		{"no THEN", "## s\n\nGIVEN model:\n\n    (empty)\n", "at least one THEN"},
		{"bad action", "## s\n\nGIVEN model:\n\n    (empty)\n\nWHEN:\n\n    frobnicate\n\nTHEN page:\n\n    >\n", "unknown action"},
		{"bad check", "## s\n\nGIVEN model:\n\n    (empty)\n\nTHEN check:\n\n    the moon is full\n", "unknown check"},
		{"bad model line", "## s\n\nGIVEN model:\n\n    [y] weird\n\nTHEN page:\n\n    >\n", "bad model line"},
		{"duplicate titles", "## s\n\nGIVEN model:\n\n    [ ] a\n    [x] a\n\nTHEN page:\n\n    >\n", "duplicate title"},
		{"stray prose in scenario", "## s\n\nGIVEN model:\n\n    (empty)\n\nsome stray prose\n\nTHEN page:\n\n    >\n", "unrecognized line"},
		{"duplicate section", "## s\n\nGIVEN model:\n\n    (empty)\n\nGIVEN model:\n\n    (empty)\n\nTHEN page:\n\n    >\n", "duplicate section"},
		{"generic click not in vocabulary", "## s\n\nGIVEN model:\n\n    (empty)\n\nWHEN:\n\n    click \"some button\"\n\nTHEN page:\n\n    >\n", "unknown action"},
		{"bad route", "## s\n\nGIVEN model:\n\n    (empty)\n\nGIVEN route: active\n\nTHEN page:\n\n    >\n", "route must start"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := parseString(t, c.doc)
			if err == nil || !strings.Contains(err.Error(), c.wantErr) {
				t.Errorf("err = %v, want containing %q", err, c.wantErr)
			}
		})
	}
}

func TestParseNoteHandling(t *testing.T) {
	doc := `# title

Preamble prose is ignored.

## s

GIVEN model:

    (empty)

NOTE: a note that spans
multiple lines is swallowed until a blank line.

THEN page:

    >

NOTE: trailing note.
`
	scs, err := parseString(t, doc)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(scs) != 1 || !reflect.DeepEqual(scs[0].ThenPage, []string{">"}) {
		t.Errorf("got %+v", scs)
	}
}

func TestModelNotationRoundTrip(t *testing.T) {
	lines := []string{"#a1 [ ] buy milk", "[x] walk the dog"}
	m, err := ParseModel(lines)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(m.Notation(), lines) {
		t.Errorf("Notation() = %q, want %q", m.Notation(), lines)
	}
	empty, err := ParseModel([]string{"(empty)"})
	if err != nil {
		t.Fatal(err)
	}
	if len(empty) != 0 || empty == nil {
		t.Errorf("empty model = %#v, want non-nil empty", empty)
	}
	if !reflect.DeepEqual(empty.Notation(), []string{"(empty)"}) {
		t.Errorf("empty Notation() = %q", empty.Notation())
	}
}
