# Application Specification

This is a TODO application.


## Main screen

Use the @main-screen-template.html and .css


## Functionality

### No todos

When there are no todos, `#main` and `#footer` should be hidden.

### New todo

New todos are entered in the input at the top of the app. The input element should be focused when the page is loaded, preferably by using the `autofocus` input attribute. Pressing Enter creates the todo, appends it to the todo list, and clears the input. Make sure to `.trim()` the input and then check that it's not empty before creating a new todo.

### Mark all as complete

This checkbox toggles all the todos to the same state as itself. Make sure to clear the checked state after the "Clear completed" button is clicked. The "Mark all as complete" checkbox should also be updated when single todo items are checked/unchecked. Eg. When all the todos are checked it should also get checked.

### Item

Each todo item's `<li>` element carries a `data-id` attribute equal to the item's persisted `id`.

A todo item has three possible interactions:

1. Clicking the checkbox marks the todo as complete by updating its `completed` value and toggling the class `completed` on its parent `<li>`

2. Double-clicking the `<label>` activates editing mode, by toggling the `.editing` class on its `<li>`

3. Hovering over the todo shows the remove button (`.destroy`)

### Editing

When editing mode is activated it will hide the other controls and bring forward an input that contains the todo title, which should be focused (`.focus()`). The edit should be saved on both blur and enter, and the `editing` class should be removed. Make sure to `.trim()` the input and then check that it's not empty. If it's empty the todo should instead be destroyed. If escape is pressed during the edit, the edit state should be left and any changes be discarded.

### Counter

Displays the number of active todos in a pluralized form. Make sure the number is wrapped by a `<strong>` tag. Also make sure to pluralize the `item` word correctly: `0 items`, `1 item`, `2 items`. Example: **2** items left

### Clear completed button

Removes completed todos when clicked. Should be hidden when there are no completed todos.

### Persistence

Your app should dynamically persist the todos, immediately after every interaction. Use the keys `id`, `title`, `completed` for each item. Each item's `id` is an opaque non-empty string, assigned when the todo is created and never changed afterwards (editing a todo's title must not change its `id`). The app must accept any non-empty string as an id in previously persisted data. Editing mode should not be persisted.

### Routing

The following routes should be implemented: `#/` (all - default), `#/active` and `#/completed` (`#!/` is also allowed). When the route changes, the todo list should be filtered on a model level and the `selected` class on the filter links should be toggled. When an item is updated while in a filtered state, it should be updated accordingly. E.g. if the filter is `Active` and the item is checked, it should be hidden. Make sure the active filter is persisted on reload.


## Technical constraints

### README

The app must include a README describing the the general implementation, and the build process.

### Code

Please try to keep the HTML as close to the template as possible. Remove the comments from the HTML when you are finished.


