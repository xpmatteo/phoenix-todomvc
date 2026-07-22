package browser

import (
	"fmt"

	"todomvc/evals/runner/internal/scenario"
)

// Check evaluates one THEN check. It returns ok plus a short description of
// what was actually observed, for failure messages.
func (s *Session) Check(c scenario.Check) (bool, string, error) {
	switch c.Kind {
	case scenario.CheckFocusNewTodo:
		return s.boolCheck(`(() => {
` + jsHelpers + `
			const el = document.querySelector('.new-todo');
			if (!el) return [false, 'no .new-todo input on the page'];
			if (document.activeElement !== el) {
				const a = document.activeElement;
				return [false, 'focus is on ' +
					(a ? a.tagName + '.' + a.className : 'nothing')];
			}
			return [true, ''];
		})()`)
	case scenario.CheckFocusEditField:
		return s.boolCheck(fmt.Sprintf(`(() => {
`+jsHelpers+`
			const m = rowsByTitle(%q);
			if (m.length !== 1) return [false, m.length +
				' displayed todos match that title, want exactly 1'];
			const edit = m[0].querySelector('.edit');
			if (!edit) return [false, 'the todo has no .edit field'];
			if (document.activeElement !== edit) {
				const a = document.activeElement;
				return [false, 'focus is on ' +
					(a ? a.tagName + '.' + a.className : 'nothing')];
			}
			return [true, ''];
		})()`, c.Title))
	case scenario.CheckDestroyVisible, scenario.CheckDestroyHidden:
		wantVisible := c.Kind == scenario.CheckDestroyVisible
		return s.boolCheck(fmt.Sprintf(`(() => {
`+jsHelpers+`
			const m = rowsByTitle(%q);
			if (m.length !== 1) return [false, m.length +
				' displayed todos match that title, want exactly 1'];
			const d = m[0].querySelector('.destroy');
			const visible = !!(d && vis(d));
			if (visible !== %t) return [false,
				'destroy button is ' + (visible ? 'visible' : 'hidden')];
			return [true, ''];
		})()`, c.Title, wantVisible))
	default:
		return false, "", fmt.Errorf("unimplemented check kind %q", c.Kind)
	}
}

// boolCheck evaluates an expression returning [ok, detail].
func (s *Session) boolCheck(expr string) (bool, string, error) {
	var res []interface{}
	if err := s.eval(expr, &res); err != nil {
		return false, "", err
	}
	if len(res) != 2 {
		return false, "", fmt.Errorf("check expression returned %v", res)
	}
	ok, _ := res[0].(bool)
	detail, _ := res[1].(string)
	return ok, detail, nil
}
