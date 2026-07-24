# Application Specification

This is a TODO application.


## Main screen

Use the @main-screen-template.html and .css


## Functionality

### No todos

When there are no todos, `#main` and `#footer` should be hidden.

### New todo

New todos are entered in the input at the top of the app. Focus that input when the page loads, preferably with the `autofocus` attribute. Pressing Enter creates the todo, appends it to the list, and clears the input. Before creating a todo, `.trim()` the input and check it is not empty.

### Mark all as complete

This checkbox toggles all todos to its own state. Clear its checked state after the "Clear completed" button is clicked. Update it when single todo items are checked or unchecked: when all todos are checked, it becomes checked too.

### Item

Each todo item's `<li>` element carries a `data-id` attribute equal to the item's persisted `id`.

A todo item has three possible interactions:

1. Clicking the checkbox marks the todo as complete by updating its `completed` value and toggling the class `completed` on its parent `<li>`

2. Double-clicking the `<label>` activates editing mode, by toggling the `.editing` class on its `<li>`

3. Hovering over the todo shows the remove button (`.destroy`)

### Editing

Activating editing mode hides the other controls and brings forward an input containing the todo title. Focus that input (`.focus()`). Blur and Enter both save the edit and remove the `editing` class. `.trim()` the input and check it is not empty; if empty, destroy the todo instead. Escape leaves editing mode and discards any changes.

### Counter

Displays the number of active todos, pluralized. Wrap the number in a `<strong>` tag. Pluralize `item` correctly: `0 items`, `1 item`, `2 items`. Example: **2** items left

### Clear completed button

Removes completed todos when clicked. Should be hidden when there are no completed todos.

### Persistence

Persist the todos immediately after every interaction. Use the keys `id`, `title`, `completed` for each item. Each item's `id` is an opaque non-empty string, assigned when the todo is created and never changed afterwards (editing a todo's title must not change its `id`). The app must accept any non-empty string as an id in previously persisted data. Editing mode should not be persisted.

### Routing

The app is server-rendered: each route is a distinct URL path rendered by the server. Implement `/` (all - default), `/active` and `/completed`. The filter links in the footer navigate to these routes, and the `selected` class is set on the link matching the current route. The displayed todo list contains only the items matching the route's filter. When an item is updated while in a filtered state, the displayed list updates accordingly: e.g. if the filter is `Active` and the item is checked, it disappears from the list. Reloading the page keeps the current filter.


The spec deliberately ends here: it contains only behavior observable through the eval surface. Constraints on *how* the app is built live in `spec/architecture.md`; the eval harness interface lives in `evals/HARNESS.md` (see the boundary rule in `README.md`).


