package run

import (
	"reflect"
	"strings"
	"testing"

	"todomvc/evals/runner/internal/browser"
	"todomvc/evals/runner/internal/harness"
	"todomvc/evals/runner/internal/scenario"
)

func TestEraseUnrequestedIDs(t *testing.T) {
	actual := []string{
		"v >",
		"#a1 [ ] buy milk",
		"#b2 [x] ~walk the dog~",
		"#c3 [edit: feed the cat]",
		"-- **2** items left | (All) Active Completed",
	}
	expected := []string{
		"v >",
		"#a1 [ ] buy milk",   // id requested: kept
		"[x] ~walk the dog~", // no id requested: erased
		"[edit: feed the cat]",
		"-- **2** items left | (All) Active Completed",
	}
	got := EraseUnrequestedIDs(actual, expected)
	want := []string{
		"v >",
		"#a1 [ ] buy milk",
		"[x] ~walk the dog~",
		"[edit: feed the cat]",
		"-- **2** items left | (All) Active Completed",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("EraseUnrequestedIDs = %q, want %q", got, want)
	}

	// Extra actual lines beyond the expected list keep their ids.
	got = EraseUnrequestedIDs([]string{"#x [ ] extra"}, nil)
	if !reflect.DeepEqual(got, []string{"#x [ ] extra"}) {
		t.Errorf("extra line = %q, want id kept", got)
	}
}

func TestCompareModel(t *testing.T) {
	expected := scenario.Model{
		{ID: "a1", Title: "buy milk"},
		{Title: "walk the dog", Completed: true},
	}
	ok := []harness.Item{
		{ID: "a1", Title: "buy milk"},
		{ID: "whatever", Title: "walk the dog", Completed: true},
	}
	if p := CompareModel(expected, ok); len(p) != 0 {
		t.Errorf("problems = %v, want none", p)
	}

	cases := []struct {
		name   string
		actual []harness.Item
		want   string
	}{
		{"wrong count", []harness.Item{{ID: "a1", Title: "buy milk"}}, "has 1 todos, want 2"},
		{"wrong title", []harness.Item{{ID: "a1", Title: "buy soap"}, {ID: "x", Title: "walk the dog", Completed: true}}, `title = "buy soap"`},
		{"wrong completed", []harness.Item{{ID: "a1", Title: "buy milk", Completed: true}, {ID: "x", Title: "walk the dog", Completed: true}}, "completed = true, want false"},
		{"wrong id where asserted", []harness.Item{{ID: "zz", Title: "buy milk"}, {ID: "x", Title: "walk the dog", Completed: true}}, `id = "zz", want "a1"`},
		{"wrong order", []harness.Item{{ID: "x", Title: "walk the dog", Completed: true}, {ID: "a1", Title: "buy milk"}}, "title"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			p := CompareModel(expected, c.actual)
			if len(p) == 0 || !strings.Contains(strings.Join(p, "\n"), c.want) {
				t.Errorf("problems = %v, want containing %q", p, c.want)
			}
		})
	}
}

func TestVerifyIntegrity(t *testing.T) {
	persisted := []harness.Item{
		{ID: "a1", Title: "buy milk"},
		{ID: "b2", Title: "walk the dog", Completed: true},
	}
	displayed := []browser.DisplayedItem{
		{Title: "buy milk", ID: "a1", HasID: true},
		{Title: "walk the dog", ID: "b2", HasID: true},
	}
	if p := VerifyIntegrity(persisted, displayed); len(p) != 0 {
		t.Errorf("problems = %v, want none", p)
	}

	// A filtered view (subset displayed) is fine.
	if p := VerifyIntegrity(persisted, displayed[:1]); len(p) != 0 {
		t.Errorf("filtered: problems = %v, want none", p)
	}

	// Swapped ids between two rows must fail.
	swapped := []browser.DisplayedItem{
		{Title: "buy milk", ID: "b2", HasID: true},
		{Title: "walk the dog", ID: "a1", HasID: true},
	}
	if p := VerifyIntegrity(persisted, swapped); len(p) != 2 {
		t.Errorf("swapped ids: problems = %v, want 2", p)
	}

	// Empty persisted id, missing data-id, phantom displayed row.
	p := VerifyIntegrity(
		[]harness.Item{{ID: "", Title: "buy milk"}},
		[]browser.DisplayedItem{
			{Title: "buy milk", HasID: false},
			{Title: "ghost", ID: "g", HasID: true},
		})
	joined := strings.Join(p, "\n")
	for _, want := range []string{"empty id", "no data-id", "no persisted counterpart"} {
		if !strings.Contains(joined, want) {
			t.Errorf("problems %v missing %q", p, want)
		}
	}
}

func TestSideBySide(t *testing.T) {
	out := SideBySide(
		[]string{"same", "left only is long enough", "old"},
		[]string{"same", "left only is long enough", "new", "extra"},
		"expected", "actual")
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) != 5 {
		t.Fatalf("got %d lines:\n%s", len(lines), out)
	}
	if !strings.Contains(lines[0], "expected") || !strings.Contains(lines[0], "actual") {
		t.Errorf("header = %q", lines[0])
	}
	if strings.HasPrefix(strings.TrimLeft(lines[1], " "), "!") {
		t.Errorf("equal row marked: %q", lines[1])
	}
	for _, i := range []int{3, 4} {
		if !strings.Contains(lines[i], "!") {
			t.Errorf("row %d not marked: %q", i, lines[i])
		}
	}
	if !strings.Contains(lines[3], "old") || !strings.Contains(lines[3], "new") {
		t.Errorf("row 3 = %q, want both sides", lines[3])
	}
}

func TestSeedItems(t *testing.T) {
	m := scenario.Model{
		{ID: "a1", Title: "one"},
		{Title: "two", Completed: true},
		{Title: "three"},
	}
	items, err := seedItems(m)
	if err != nil {
		t.Fatal(err)
	}
	if items[0].ID != "a1" {
		t.Errorf("explicit id not passed verbatim: %v", items[0])
	}
	seen := map[string]bool{}
	for _, it := range items {
		if it.ID == "" {
			t.Errorf("item %q has empty id", it.Title)
		}
		if seen[it.ID] {
			t.Errorf("duplicate id %q", it.ID)
		}
		seen[it.ID] = true
	}
	if items[1].Title != "two" || !items[1].Completed {
		t.Errorf("item state lost: %v", items[1])
	}

	empty, err := seedItems(scenario.Model{})
	if err != nil || len(empty) != 0 {
		t.Errorf("empty model: %v, %v", empty, err)
	}
}
