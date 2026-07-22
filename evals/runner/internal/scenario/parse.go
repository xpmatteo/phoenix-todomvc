package scenario

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// ParseDir parses every *.md file in dir, in lexical filename order.
func ParseDir(dir string) ([]Scenario, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var names []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)
	if len(names) == 0 {
		return nil, fmt.Errorf("no scenario files (*.md) in %s", dir)
	}
	var all []Scenario
	for _, n := range names {
		path := filepath.Join(dir, n)
		f, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		scs, err := Parse(path, f)
		f.Close()
		if err != nil {
			return nil, err
		}
		all = append(all, scs...)
	}
	return all, nil
}

type section int

const (
	secNone section = iota
	secGivenModel
	secWhen
	secThenPage
	secThenModel
	secThenCheck
)

type rawScenario struct {
	name string
	line int

	route    string
	hasRoute bool

	blocks map[section][]string
	seen   map[section]bool
}

// Parse parses one scenario file. The name is used in error messages and
// recorded on each Scenario.
func Parse(name string, r io.Reader) ([]Scenario, error) {
	var raws []*rawScenario
	var cur *rawScenario
	var curSec section
	inNote := false

	scanner := bufio.NewScanner(r)
	lineNo := 0
	for scanner.Scan() {
		lineNo++
		line := strings.TrimRight(scanner.Text(), " \t\r")
		trimmed := strings.TrimSpace(line)

		if trimmed == "" {
			inNote = false
			continue
		}
		if strings.HasPrefix(trimmed, "NOTE:") {
			inNote = true
			continue
		}
		if inNote {
			// NOTE commentary continues until a blank line.
			continue
		}
		if strings.HasPrefix(line, "## ") {
			cur = &rawScenario{
				name:   strings.TrimSpace(line[3:]),
				line:   lineNo,
				blocks: map[section][]string{},
				seen:   map[section]bool{},
			}
			raws = append(raws, cur)
			curSec = secNone
			continue
		}
		if strings.HasPrefix(line, "# ") {
			// File title; anything before the first "## " is preamble.
			continue
		}
		if cur == nil {
			// Prose preamble before the first scenario heading.
			continue
		}
		if strings.HasPrefix(line, "    ") {
			if curSec == secNone {
				return nil, fmt.Errorf("%s:%d: indented line outside any section", name, lineNo)
			}
			cur.blocks[curSec] = append(cur.blocks[curSec], line[4:])
			continue
		}
		// Unindented line inside a scenario: must be a section keyword.
		sec, route, ok := parseKeyword(line)
		if !ok {
			return nil, fmt.Errorf("%s:%d: unrecognized line %q", name, lineNo, line)
		}
		if sec == secNone { // GIVEN route:
			if cur.hasRoute {
				return nil, fmt.Errorf("%s:%d: duplicate GIVEN route:", name, lineNo)
			}
			if !strings.HasPrefix(route, "/") {
				return nil, fmt.Errorf("%s:%d: GIVEN route must start with %q, got %q", name, lineNo, "/", route)
			}
			cur.route = route
			cur.hasRoute = true
			continue
		}
		if cur.seen[sec] {
			return nil, fmt.Errorf("%s:%d: duplicate section %q", name, lineNo, line)
		}
		cur.seen[sec] = true
		curSec = sec
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if len(raws) == 0 {
		return nil, fmt.Errorf("%s: no scenarios (no %q headings)", name, "## ")
	}

	var out []Scenario
	for _, rw := range raws {
		sc, err := buildScenario(name, rw)
		if err != nil {
			return nil, err
		}
		out = append(out, *sc)
	}
	return out, nil
}

// parseKeyword recognizes a section keyword line. For "GIVEN route:" it
// returns secNone with the route value and ok=true.
func parseKeyword(line string) (section, string, bool) {
	switch line {
	case "GIVEN model:":
		return secGivenModel, "", true
	case "WHEN:":
		return secWhen, "", true
	case "THEN page:":
		return secThenPage, "", true
	case "THEN model:":
		return secThenModel, "", true
	case "THEN check:":
		return secThenCheck, "", true
	}
	if rest, ok := strings.CutPrefix(line, "GIVEN route:"); ok {
		return secNone, strings.TrimSpace(rest), true
	}
	return secNone, "", false
}

func buildScenario(file string, rw *rawScenario) (*Scenario, error) {
	where := fmt.Sprintf("%s:%d: scenario %q", file, rw.line, rw.name)
	sc := &Scenario{File: file, Name: rw.name, Line: rw.line, Route: "/"}
	if rw.hasRoute {
		sc.Route = rw.route
	}

	if !rw.seen[secGivenModel] {
		return nil, fmt.Errorf("%s: missing required GIVEN model:", where)
	}
	given, err := ParseModel(rw.blocks[secGivenModel])
	if err != nil {
		return nil, fmt.Errorf("%s: GIVEN model: %w", where, err)
	}
	sc.Given = given
	if err := checkUnique(given, true); err != nil {
		return nil, fmt.Errorf("%s: GIVEN model: %w", where, err)
	}

	if rw.seen[secWhen] {
		for _, l := range rw.blocks[secWhen] {
			a, err := ParseAction(l)
			if err != nil {
				return nil, fmt.Errorf("%s: WHEN: %w", where, err)
			}
			sc.When = append(sc.When, a)
		}
		if len(sc.When) == 0 {
			return nil, fmt.Errorf("%s: WHEN: section is empty", where)
		}
	}

	if rw.seen[secThenPage] {
		sc.HasThenPage = true
		sc.ThenPage = rw.blocks[secThenPage]
		if len(sc.ThenPage) == 0 {
			return nil, fmt.Errorf("%s: THEN page: section is empty", where)
		}
	}
	if rw.seen[secThenModel] {
		sc.HasThenModel = true
		m, err := ParseModel(rw.blocks[secThenModel])
		if err != nil {
			return nil, fmt.Errorf("%s: THEN model: %w", where, err)
		}
		if err := checkUnique(m, false); err != nil {
			return nil, fmt.Errorf("%s: THEN model: %w", where, err)
		}
		sc.ThenModel = m
	}
	if rw.seen[secThenCheck] {
		for _, l := range rw.blocks[secThenCheck] {
			c, err := ParseCheck(l)
			if err != nil {
				return nil, fmt.Errorf("%s: THEN check: %w", where, err)
			}
			sc.Checks = append(sc.Checks, c)
		}
		if len(sc.Checks) == 0 {
			return nil, fmt.Errorf("%s: THEN check: section is empty", where)
		}
	}

	if !sc.HasThenPage && !sc.HasThenModel && len(sc.Checks) == 0 {
		return nil, fmt.Errorf("%s: at least one THEN section is required", where)
	}
	return sc, nil
}

var todoLineRe = regexp.MustCompile(`^(?:#(\S+) )?\[( |x)\] (.+)$`)

// ParseModel parses a model notation block: one todo per line, or "(empty)".
func ParseModel(lines []string) (Model, error) {
	if len(lines) == 0 {
		return nil, fmt.Errorf("model block is empty (use %q for no todos)", "(empty)")
	}
	if len(lines) == 1 && lines[0] == "(empty)" {
		return Model{}, nil
	}
	m := make(Model, 0, len(lines))
	for _, l := range lines {
		g := todoLineRe.FindStringSubmatch(l)
		if g == nil {
			return nil, fmt.Errorf("bad model line %q", l)
		}
		m = append(m, Todo{ID: g[1], Title: g[3], Completed: g[2] == "x"})
	}
	return m, nil
}

func checkUnique(m Model, checkIDs bool) error {
	titles := map[string]bool{}
	ids := map[string]bool{}
	for _, t := range m {
		if titles[t.Title] {
			return fmt.Errorf("duplicate title %q (titles must be unique within a scenario)", t.Title)
		}
		titles[t.Title] = true
		if checkIDs && t.ID != "" {
			if ids[t.ID] {
				return fmt.Errorf("duplicate id %q", t.ID)
			}
			ids[t.ID] = true
		}
	}
	return nil
}

var (
	reType    = regexp.MustCompile(`^type "(.*)"$`)
	rePress   = regexp.MustCompile(`^press (Enter|Escape)$`)
	reToggle  = regexp.MustCompile(`^click toggle of "(.+)"$`)
	reDestroy = regexp.MustCompile(`^click destroy of "(.+)"$`)
	reDbl     = regexp.MustCompile(`^dblclick "(.+)"$`)
	reFilter  = regexp.MustCompile(`^click filter "(All|Active|Completed)"$`)
	reGoTo    = regexp.MustCompile(`^go to "(/[^"]*)"$`)
	reHover   = regexp.MustCompile(`^hover "(.+)"$`)
)

// ParseAction parses one WHEN line against the closed action vocabulary.
func ParseAction(line string) (Action, error) {
	a := Action{Raw: line}
	switch line {
	case "clear":
		a.Verb = VerbClear
		return a, nil
	case "blur":
		a.Verb = VerbBlur
		return a, nil
	case "reload":
		a.Verb = VerbReload
		return a, nil
	case "click toggle-all":
		a.Verb = VerbToggleAll
		return a, nil
	case `click "Clear completed"`:
		a.Verb = VerbClearCompleted
		return a, nil
	}
	type rule struct {
		re *regexp.Regexp
		v  Verb
	}
	for _, r := range []rule{
		{reType, VerbType},
		{reToggle, VerbToggle},
		{reDestroy, VerbDestroy},
		{reDbl, VerbDblclick},
		{reFilter, VerbFilter},
		{reGoTo, VerbGoTo},
		{reHover, VerbHover},
	} {
		if g := r.re.FindStringSubmatch(line); g != nil {
			a.Verb = r.v
			a.Arg = g[1]
			return a, nil
		}
	}
	if g := rePress.FindStringSubmatch(line); g != nil {
		if g[1] == "Enter" {
			a.Verb = VerbPressEnter
		} else {
			a.Verb = VerbPressEscape
		}
		return a, nil
	}
	return a, fmt.Errorf("unknown action %q (closed vocabulary, see DSL.md)", line)
}

var (
	reCheckEdit    = regexp.MustCompile(`^focus is in the edit field of "(.+)"$`)
	reCheckDestroy = regexp.MustCompile(`^destroy button of "(.+)" is (visible|hidden)$`)
)

// ParseCheck parses one THEN check line against the closed check registry.
func ParseCheck(line string) (Check, error) {
	c := Check{Raw: line}
	if line == "focus is on the new-todo input" {
		c.Kind = CheckFocusNewTodo
		return c, nil
	}
	if g := reCheckEdit.FindStringSubmatch(line); g != nil {
		c.Kind = CheckFocusEditField
		c.Title = g[1]
		return c, nil
	}
	if g := reCheckDestroy.FindStringSubmatch(line); g != nil {
		if g[2] == "visible" {
			c.Kind = CheckDestroyVisible
		} else {
			c.Kind = CheckDestroyHidden
		}
		c.Title = g[1]
		return c, nil
	}
	return c, fmt.Errorf("unknown check %q (closed registry, see DSL.md)", line)
}
