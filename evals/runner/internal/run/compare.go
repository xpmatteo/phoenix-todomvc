package run

import (
	"fmt"
	"regexp"
	"strings"

	"todomvc/evals/runner/internal/browser"
	"todomvc/evals/runner/internal/harness"
	"todomvc/evals/runner/internal/scenario"
)

var leadingIDRe = regexp.MustCompile(`^#\S+ `)

// EraseUnrequestedIDs strips the leading #id from actual projection lines
// whose expected counterpart (same position) carries none, so ids stay
// opt-in per line while the diff remains exact (DSL.md § Runner obligations).
func EraseUnrequestedIDs(actual, expected []string) []string {
	out := make([]string, len(actual))
	for i, a := range actual {
		if i < len(expected) && !strings.HasPrefix(expected[i], "#") {
			a = leadingIDRe.ReplaceAllString(a, "")
		}
		out[i] = a
	}
	return out
}

func linesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// CompareModel checks the persisted model against the expectation: title,
// completed state and order always; id only where the expected line carries
// one. It returns a list of mismatch descriptions (empty means match).
func CompareModel(expected scenario.Model, actual []harness.Item) []string {
	var problems []string
	if len(actual) != len(expected) {
		problems = append(problems, fmt.Sprintf(
			"persisted model has %d todos, want %d", len(actual), len(expected)))
	}
	n := min(len(actual), len(expected))
	for i := 0; i < n; i++ {
		e, a := expected[i], actual[i]
		if a.Title != e.Title {
			problems = append(problems, fmt.Sprintf(
				"todo %d: title = %q, want %q", i+1, a.Title, e.Title))
		}
		if a.Completed != e.Completed {
			problems = append(problems, fmt.Sprintf(
				"todo %d (%q): completed = %t, want %t", i+1, a.Title, a.Completed, e.Completed))
		}
		if e.ID != "" && a.ID != e.ID {
			problems = append(problems, fmt.Sprintf(
				"todo %d (%q): id = %q, want %q", i+1, a.Title, a.ID, e.ID))
		}
	}
	return problems
}

// VerifyIntegrity enforces the invariants the runner checks on every
// scenario (DSL.md § Runner obligations): every persisted item has a
// non-empty string id, and every displayed row's data-id equals the id of
// the persisted item bearing the same title.
func VerifyIntegrity(persisted []harness.Item, displayed []browser.DisplayedItem) []string {
	var problems []string
	byTitle := map[string]harness.Item{}
	for i, it := range persisted {
		if it.ID == "" {
			problems = append(problems, fmt.Sprintf(
				"persisted todo %d (%q) has an empty id", i+1, it.Title))
		}
		byTitle[it.Title] = it
	}
	for _, d := range displayed {
		p, found := byTitle[d.Title]
		switch {
		case !found:
			problems = append(problems, fmt.Sprintf(
				"displayed todo %q has no persisted counterpart with that title", d.Title))
		case !d.HasID:
			problems = append(problems, fmt.Sprintf(
				"displayed todo %q has no data-id attribute (want %q)", d.Title, p.ID))
		case d.ID != p.ID:
			problems = append(problems, fmt.Sprintf(
				"displayed todo %q has data-id %q, but the persisted item with that title has id %q",
				d.Title, d.ID, p.ID))
		}
	}
	return problems
}

// ItemsToModel converts persisted items to model values for rendering.
func ItemsToModel(items []harness.Item) scenario.Model {
	m := make(scenario.Model, 0, len(items))
	for _, it := range items {
		m = append(m, scenario.Todo{ID: it.ID, Title: it.Title, Completed: it.Completed})
	}
	return m
}

// SideBySide renders expected and actual line lists in two columns, with a
// "!" marker on rows that differ.
func SideBySide(expected, actual []string, leftHeader, rightHeader string) string {
	width := len(leftHeader)
	for _, l := range expected {
		if len(l) > width {
			width = len(l)
		}
	}
	var b strings.Builder
	fmt.Fprintf(&b, "    %-*s | %s\n", width, leftHeader, rightHeader)
	n := max(len(expected), len(actual))
	for i := 0; i < n; i++ {
		var e, a string
		if i < len(expected) {
			e = expected[i]
		}
		if i < len(actual) {
			a = actual[i]
		}
		marker := "  "
		if e != a {
			marker = "! "
		}
		fmt.Fprintf(&b, "  %s%-*s | %s\n", marker, width, e, a)
	}
	return b.String()
}
