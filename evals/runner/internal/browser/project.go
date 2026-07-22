package browser

import "strings"

// jsHelpers defines visibility and item lookup shared by projection, actions
// and checks. Visibility means rendered visibility per HARNESS.md: computed
// style, not mere DOM presence.
const jsHelpers = `
	const vis = (el) => {
		if (!el) return false;
		const st = getComputedStyle(el);
		if (st.visibility === 'hidden' || st.visibility === 'collapse') return false;
		for (let e = el; e; e = e.parentElement) {
			if (getComputedStyle(e).display === 'none') return false;
		}
		return true;
	};
	// Displayed todo rows, in displayed order.
	const rows = () =>
		[...document.querySelectorAll('.todo-list > li')].filter(vis);
	// The label text is the todo's current title (DSL: exact current label text).
	const rowTitle = (li) => {
		const label = li.querySelector('label');
		return label ? label.textContent.trim() : null;
	};
	// Displayed rows whose title matches exactly. The row itself must be
	// displayed; the label may be hidden (editing mode) since textContent
	// is unaffected by visibility.
	const rowsByTitle = (title) => rows().filter(li => rowTitle(li) === title);
`

// jsProject computes the ASCII page projection per DSL.md § Page projection.
// Item lines always carry their data-id (when present); the runner erases ids
// the expected projection does not ask for.
const jsProject = `(() => {
` + jsHelpers + `
	const lines = [];

	// 1. Input row.
	const newTodo = document.querySelector('.new-todo');
	if (!newTodo || !vis(newTodo)) {
		lines.push('(no new-todo input)');
	} else {
		const main = document.querySelector('.main');
		let prefix = '';
		if (main && vis(main)) {
			const ta = document.querySelector('.toggle-all');
			prefix = (ta && ta.checked) ? '(v) ' : 'v ';
		}
		const v = newTodo.value;
		lines.push(prefix + '>' + (v !== '' ? ' ' + v : ''));
	}

	// 2. Todo items. For each displayed row: a normal line if its view
	// controls are rendered, an [edit: …] line if its edit field is rendered.
	// A correct app renders exactly one of the two; emitting whatever is
	// actually visible lets the diff expose an app that shows both.
	for (const li of rows()) {
		const id = li.getAttribute('data-id');
		const idPrefix = id ? '#' + id + ' ' : '';
		const toggle = li.querySelector('.toggle');
		const label = li.querySelector('label');
		const destroy = li.querySelector('.destroy');
		const viewShown = (toggle && vis(toggle)) || (label && vis(label)) ||
			(destroy && vis(destroy));
		if (viewShown) {
			const title = label ? label.textContent.trim() : '(no label)';
			const checked = !!(toggle && toggle.checked);
			// Completed *styling* is judged by the rendered look: the
			// line-through the template CSS applies via li.completed.
			const struck = !!(label && getComputedStyle(label)
				.textDecorationLine.includes('line-through'));
			lines.push(idPrefix + (checked ? '[x] ' : '[ ] ') +
				(struck ? '~' + title + '~' : title));
		}
		const edit = li.querySelector('.edit');
		if (edit && vis(edit)) {
			lines.push(idPrefix + '[edit: ' + edit.value + ']');
		}
	}

	// 3. Footer.
	const footer = document.querySelector('.footer');
	if (footer && vis(footer)) {
		let counter = '';
		const tc = footer.querySelector('.todo-count');
		if (tc && vis(tc)) {
			for (const n of tc.childNodes) {
				if (n.nodeType === 1 && n.tagName === 'STRONG') {
					counter += '**' + n.textContent + '**';
				} else {
					counter += n.textContent;
				}
			}
		}
		let line = '-- ' + counter.trim();
		const filters = [...footer.querySelectorAll('.filters a')].filter(vis)
			.map(a => {
				const t = a.textContent.trim();
				return a.classList.contains('selected') ? '(' + t + ')' : t;
			});
		if (filters.length) line += ' | ' + filters.join(' ');
		const cc = footer.querySelector('.clear-completed');
		if (cc && vis(cc)) line += ' | [Clear completed]';
		lines.push(line);
	}
	return lines;
})()`

// Project returns the current page projection, top to bottom. Trailing
// whitespace is trimmed from every line, mirroring what markdown scenario
// files can reliably express in expected projections.
func (s *Session) Project() ([]string, error) {
	var lines []string
	if err := s.eval(jsProject, &lines); err != nil {
		return nil, err
	}
	for i, l := range lines {
		lines[i] = strings.TrimRight(l, " \t")
	}
	return lines, nil
}

// DisplayedItem is a rendered todo row as needed for the id-integrity
// verification of DSL.md § Runner obligations.
type DisplayedItem struct {
	Title string `json:"title"`
	ID    string `json:"id"`
	HasID bool   `json:"hasId"`
}

const jsDisplayedItems = `(() => {
` + jsHelpers + `
	return rows().map(li => ({
		title: rowTitle(li) ?? '',
		id: li.getAttribute('data-id') ?? '',
		hasId: li.hasAttribute('data-id'),
	}));
})()`

// DisplayedItems lists the displayed todo rows with their data-id attributes.
func (s *Session) DisplayedItems() ([]DisplayedItem, error) {
	var items []DisplayedItem
	if err := s.eval(jsDisplayedItems, &items); err != nil {
		return nil, err
	}
	return items, nil
}
