// Package scenario parses the eval DSL defined in evals/DSL.md.
package scenario

import "fmt"

// Todo is one line of model notation. An empty ID means the line carried no #id.
type Todo struct {
	ID        string
	Title     string
	Completed bool
}

// Model is an ordered list of todos. A non-nil empty Model is "(empty)".
type Model []Todo

// Notation renders the model back into DSL model notation, one line per todo.
func (m Model) Notation() []string {
	if len(m) == 0 {
		return []string{"(empty)"}
	}
	lines := make([]string, 0, len(m))
	for _, t := range m {
		box := "[ ]"
		if t.Completed {
			box = "[x]"
		}
		if t.ID != "" {
			lines = append(lines, fmt.Sprintf("#%s %s %s", t.ID, box, t.Title))
		} else {
			lines = append(lines, fmt.Sprintf("%s %s", box, t.Title))
		}
	}
	return lines
}

// Verb identifies an action from the closed WHEN vocabulary.
type Verb string

const (
	VerbType           Verb = "type"          // Arg: text
	VerbPressEnter     Verb = "press-enter"   //
	VerbPressEscape    Verb = "press-escape"  //
	VerbClear          Verb = "clear"         //
	VerbBlur           Verb = "blur"          //
	VerbToggle         Verb = "click-toggle"  // Arg: title
	VerbDestroy        Verb = "click-destroy" // Arg: title
	VerbDblclick       Verb = "dblclick"      // Arg: title
	VerbToggleAll      Verb = "click-toggle-all"
	VerbClearCompleted Verb = "click-clear-completed"
	VerbFilter         Verb = "click-filter" // Arg: All | Active | Completed
	VerbGoTo           Verb = "go-to"        // Arg: URL path
	VerbReload         Verb = "reload"       //
	VerbHover          Verb = "hover"        // Arg: title
)

// Action is one parsed WHEN step.
type Action struct {
	Verb Verb
	Arg  string
	Raw  string
}

// CheckKind identifies a check from the closed THEN check registry.
type CheckKind string

const (
	CheckFocusNewTodo   CheckKind = "focus-new-todo"
	CheckFocusEditField CheckKind = "focus-edit-field" // Title set
	CheckDestroyVisible CheckKind = "destroy-visible"  // Title set
	CheckDestroyHidden  CheckKind = "destroy-hidden"   // Title set
)

// Check is one parsed THEN check line.
type Check struct {
	Kind  CheckKind
	Title string
	Raw   string
}

// Scenario is one "## heading" block of a scenario file.
type Scenario struct {
	File string // path of the scenario file
	Name string // the ## heading text
	Line int    // 1-based line of the heading

	Given Model
	Route string // default "/"

	When []Action

	HasThenPage  bool
	ThenPage     []string // expected projection lines, verbatim
	HasThenModel bool
	ThenModel    Model
	Checks       []Check
}
