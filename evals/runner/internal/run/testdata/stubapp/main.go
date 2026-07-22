// Stub app for testing the runner's orchestration path. It is deliberately
// NOT a TodoMVC implementation: it has no behavior at all. It renders the
// model.json file (written by the seed command, read back by the read
// command) into a template-shaped page, read-only, so the runner's
// start/poll/seed/navigate/project/read pipeline can be exercised
// hermetically. UI interactions never change anything here.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html"
	"net/http"
	"os"
	"strings"
)

type item struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Completed bool   `json:"completed"`
}

func main() {
	port := flag.String("port", "8080", "port to listen on")
	flag.Parse()
	http.HandleFunc("/", handle)
	if err := http.ListenAndServe("127.0.0.1:"+*port, nil); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

const css = `<style>
.destroy { display: none; }
li:hover .destroy { display: inline; }
li.completed label { text-decoration: line-through; }
li .edit { display: none; }
</style>`

func handle(w http.ResponseWriter, r *http.Request) {
	filter := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/"), "/")
	if filter != "" && filter != "active" && filter != "completed" {
		http.NotFound(w, r)
		return
	}

	var items []item
	if data, err := os.ReadFile("model.json"); err == nil {
		json.Unmarshal(data, &items)
	}

	var b strings.Builder
	b.WriteString("<!doctype html><html><head><meta charset=\"utf-8\">")
	b.WriteString("<title>stub</title>")
	b.WriteString(css)
	b.WriteString("</head><body><section class=\"todoapp\">")
	b.WriteString("<header class=\"header\"><h1>todos</h1>")
	b.WriteString("<input class=\"new-todo\" placeholder=\"What needs to be done?\" autofocus>")
	b.WriteString("</header>")

	if len(items) > 0 {
		active := 0
		completed := 0
		for _, it := range items {
			if it.Completed {
				completed++
			} else {
				active++
			}
		}
		checked := ""
		if active == 0 {
			checked = " checked"
		}
		fmt.Fprintf(&b, "<section class=\"main\">"+
			"<input id=\"toggle-all\" class=\"toggle-all\" type=\"checkbox\"%s>"+
			"<label for=\"toggle-all\">Mark all as complete</label><ul class=\"todo-list\">", checked)
		for _, it := range items {
			if filter == "active" && it.Completed || filter == "completed" && !it.Completed {
				continue
			}
			cls, chk := "", ""
			if it.Completed {
				cls, chk = " class=\"completed\"", " checked"
			}
			title := html.EscapeString(it.Title)
			fmt.Fprintf(&b, "<li%s data-id=\"%s\"><div class=\"view\">"+
				"<input class=\"toggle\" type=\"checkbox\"%s>"+
				"<label>%s</label><button class=\"destroy\"></button></div>"+
				"<input class=\"edit\" value=\"%s\"></li>",
				cls, html.EscapeString(it.ID), chk, title, title)
		}
		b.WriteString("</ul></section>")

		unit := "items"
		if active == 1 {
			unit = "item"
		}
		fmt.Fprintf(&b, "<footer class=\"footer\">"+
			"<span class=\"todo-count\"><strong>%d</strong> %s left</span><ul class=\"filters\">", active, unit)
		for _, f := range []struct{ href, name, key string }{
			{"/", "All", ""}, {"/active", "Active", "active"}, {"/completed", "Completed", "completed"},
		} {
			sel := ""
			if f.key == filter {
				sel = " class=\"selected\""
			}
			fmt.Fprintf(&b, "<li><a%s href=\"%s\">%s</a></li>", sel, f.href, f.name)
		}
		b.WriteString("</ul>")
		if completed > 0 {
			b.WriteString("<button class=\"clear-completed\">Clear completed</button>")
		}
		b.WriteString("</footer>")
	}
	b.WriteString("</section></body></html>")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(b.String()))
}
