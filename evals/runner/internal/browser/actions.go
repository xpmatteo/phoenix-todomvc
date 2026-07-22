package browser

import (
	"fmt"
	"strings"
	"time"

	"github.com/chromedp/cdproto/input"
	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/kb"

	"todomvc/evals/runner/internal/scenario"
)

// Do executes one WHEN action. baseURL is the app URL from the manifest,
// needed by `go to`. Every action is followed by a settle so that any
// navigation or async rendering it triggers has a chance to complete.
func (s *Session) Do(a scenario.Action, baseURL string) error {
	if err := s.dispatch(a, baseURL); err != nil {
		return fmt.Errorf("action %q: %w", a.Raw, err)
	}
	if err := s.settle(); err != nil {
		return fmt.Errorf("after action %q: %w", a.Raw, err)
	}
	return nil
}

func (s *Session) dispatch(a scenario.Action, baseURL string) error {
	switch a.Verb {
	case scenario.VerbType:
		s.awaitFocus()
		return s.run(10*time.Second, chromedp.KeyEvent(a.Arg))
	case scenario.VerbPressEnter:
		s.awaitFocus()
		return s.run(10*time.Second, chromedp.KeyEvent(kb.Enter))
	case scenario.VerbPressEscape:
		s.awaitFocus()
		return s.run(10*time.Second, chromedp.KeyEvent(kb.Escape))
	case scenario.VerbClear:
		s.awaitFocus()
		// Select the focused input's whole value, then a real Backspace key
		// deletes it natively (trusted input event, like a user would).
		var ok bool
		if err := s.eval(`(() => {
			const el = document.activeElement;
			if (!el || typeof el.select !== 'function') return false;
			el.select();
			return true;
		})()`, &ok); err != nil {
			return err
		}
		if !ok {
			return fmt.Errorf("no focused input to clear")
		}
		return s.run(10*time.Second, chromedp.KeyEvent(kb.Backspace))
	case scenario.VerbBlur:
		var ignored bool
		return s.eval(`(() => {
			if (document.activeElement) document.activeElement.blur();
			return true;
		})()`, &ignored)
	case scenario.VerbToggle:
		if err := s.markItem(a.Arg); err != nil {
			return err
		}
		return s.clickCenter("[data-eval-target] .toggle", 1)
	case scenario.VerbDestroy:
		if err := s.markItem(a.Arg); err != nil {
			return err
		}
		// The destroy button only shows on hover: hover the row first.
		if err := s.hoverCenter("[data-eval-target]"); err != nil {
			return err
		}
		return s.clickCenter("[data-eval-target] .destroy", 1)
	case scenario.VerbDblclick:
		if err := s.markItem(a.Arg); err != nil {
			return err
		}
		return s.clickCenter("[data-eval-target] label", 2)
	case scenario.VerbToggleAll:
		// The .toggle-all checkbox itself is rendered as an invisible 1px
		// control; users operate it through its label (the chevron).
		if err := s.markToggleAllLabel(); err != nil {
			return err
		}
		return s.clickCenter("[data-eval-target]", 1)
	case scenario.VerbClearCompleted:
		return s.clickCenter(".clear-completed", 1)
	case scenario.VerbFilter:
		if err := s.markFilterLink(a.Arg); err != nil {
			return err
		}
		return s.clickCenter("[data-eval-target]", 1)
	case scenario.VerbGoTo:
		return s.Navigate(strings.TrimRight(baseURL, "/") + a.Arg)
	case scenario.VerbReload:
		return s.run(20*time.Second, chromedp.Reload())
	case scenario.VerbHover:
		if err := s.markItem(a.Arg); err != nil {
			return err
		}
		return s.hoverCenter("[data-eval-target]")
	default:
		return fmt.Errorf("unimplemented verb %q", a.Verb)
	}
}

// markItem tags the displayed todo row with the given title (exact current
// label text) with data-eval-target, erroring unless exactly one matches.
func (s *Session) markItem(title string) error {
	expr := fmt.Sprintf(`(() => {
`+jsHelpers+`
		document.querySelectorAll('[data-eval-target]')
			.forEach(e => e.removeAttribute('data-eval-target'));
		const m = rowsByTitle(%q);
		if (m.length === 1) m[0].setAttribute('data-eval-target', '1');
		return m.length;
	})()`, title)
	var n int
	if err := s.eval(expr, &n); err != nil {
		return err
	}
	if n != 1 {
		return fmt.Errorf("todo %q: %d displayed todos match that title, want exactly 1", title, n)
	}
	return nil
}

func (s *Session) markToggleAllLabel() error {
	var found bool
	err := s.eval(`(() => {
		document.querySelectorAll('[data-eval-target]')
			.forEach(e => e.removeAttribute('data-eval-target'));
		const input = document.querySelector('.toggle-all');
		let label = null;
		if (input && input.id) {
			label = document.querySelector('label[for="' + input.id + '"]');
		}
		if (!label && input) label = input.nextElementSibling &&
			input.nextElementSibling.tagName === 'LABEL' ?
			input.nextElementSibling : null;
		if (!label) return false;
		label.setAttribute('data-eval-target', '1');
		return true;
	})()`, &found)
	if err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("mark-all-as-complete toggle (label of .toggle-all) not found")
	}
	return nil
}

func (s *Session) markFilterLink(name string) error {
	expr := fmt.Sprintf(`(() => {
`+jsHelpers+`
		document.querySelectorAll('[data-eval-target]')
			.forEach(e => e.removeAttribute('data-eval-target'));
		const m = [...document.querySelectorAll('.filters a')]
			.filter(a => vis(a) && a.textContent.trim() === %q);
		if (m.length === 1) m[0].setAttribute('data-eval-target', '1');
		return m.length;
	})()`, name)
	var n int
	if err := s.eval(expr, &n); err != nil {
		return err
	}
	if n != 1 {
		return fmt.Errorf("filter link %q: %d matches, want exactly 1", name, n)
	}
	return nil
}

// center scrolls the element into view and returns its viewport center.
func (s *Session) center(sel string) (x, y float64, err error) {
	expr := fmt.Sprintf(`(() => {
		const el = document.querySelector(%q);
		if (!el) return null;
		el.scrollIntoView({block: 'center', inline: 'center'});
		const r = el.getBoundingClientRect();
		return [r.left + r.width / 2, r.top + r.height / 2];
	})()`, sel)
	var pt []float64
	if err := s.eval(expr, &pt); err != nil {
		return 0, 0, err
	}
	if len(pt) != 2 {
		return 0, 0, fmt.Errorf("element %q not found", sel)
	}
	return pt[0], pt[1], nil
}

// clickCenter dispatches real CDP mouse events at the element's center:
// count=1 for a click, count=2 for a double-click.
func (s *Session) clickCenter(sel string, count int) error {
	x, y, err := s.center(sel)
	if err != nil {
		return err
	}
	acts := []chromedp.Action{input.DispatchMouseEvent(input.MouseMoved, x, y)}
	for i := 1; i <= count; i++ {
		acts = append(acts,
			input.DispatchMouseEvent(input.MousePressed, x, y).
				WithButton(input.Left).WithClickCount(int64(i)),
			input.DispatchMouseEvent(input.MouseReleased, x, y).
				WithButton(input.Left).WithClickCount(int64(i)),
		)
	}
	return s.run(10*time.Second, acts...)
}

// hoverCenter moves the pointer over the element's center, driving :hover.
func (s *Session) hoverCenter(sel string) error {
	x, y, err := s.center(sel)
	if err != nil {
		return err
	}
	return s.run(10*time.Second, input.DispatchMouseEvent(input.MouseMoved, x, y))
}

// awaitFocus gives the page a moment to place focus somewhere before a
// keyboard action: autofocus (and app-driven focus) can land a beat after
// the load event. If nothing gets focused, the action proceeds anyway — the
// keys then land nowhere and the scenario's THENs report it.
func (s *Session) awaitFocus() {
	deadline := time.Now().Add(1 * time.Second)
	for {
		var focused bool
		err := s.eval(`(() => {
			const a = document.activeElement;
			return !!(a && a !== document.body && a.tagName !== 'HTML');
		})()`, &focused)
		if err == nil && focused {
			return
		}
		if time.Now().After(deadline) {
			return
		}
		time.Sleep(25 * time.Millisecond)
	}
}

// settle waits for any navigation the previous action may have triggered to
// reach a complete document again.
func (s *Session) settle() error {
	time.Sleep(100 * time.Millisecond)
	deadline := time.Now().Add(10 * time.Second)
	for {
		var state string
		err := s.eval(`document.readyState`, &state)
		if err == nil && state == "complete" {
			return nil
		}
		if time.Now().After(deadline) {
			if err != nil {
				return fmt.Errorf("waiting for page to settle: %w", err)
			}
			return fmt.Errorf("page did not settle (readyState=%s)", state)
		}
		time.Sleep(50 * time.Millisecond)
	}
}
