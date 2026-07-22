// Package run orchestrates a whole eval run: app lifecycle per HARNESS.md,
// scenario execution, THEN evaluation, and reporting.
package run

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"strings"
	"time"

	"todomvc/evals/runner/internal/browser"
	"todomvc/evals/runner/internal/harness"
	"todomvc/evals/runner/internal/scenario"
)

// Options configures a run.
type Options struct {
	AppDir       string
	ScenariosDir string
	ChromePath   string        // optional explicit Chrome executable
	AssertWait   time.Duration // how long THEN sections are polled before failing
	Out          io.Writer
}

// Result is the outcome of one scenario.
type Result struct {
	Scenario scenario.Scenario
	Failures []string // empty means pass
}

// Run executes every scenario and reports. It returns the per-scenario
// results; err is non-nil only for setup-level problems (bad scenarios,
// app would not start, browser troubles).
func Run(ctx context.Context, opts Options) ([]Result, error) {
	if opts.AssertWait == 0 {
		opts.AssertWait = 3 * time.Second
	}
	out := opts.Out

	scenarios, err := scenario.ParseDir(opts.ScenariosDir)
	if err != nil {
		return nil, err
	}

	app, err := harness.Load(opts.AppDir)
	if err != nil {
		return nil, err
	}
	fmt.Fprintf(out, "starting app: %s (in %s)\n", app.Manifest.Start, app.Dir)
	if err := app.Start(ctx); err != nil {
		return nil, err
	}
	defer app.Stop()
	fmt.Fprintf(out, "app ready at %s; %d scenarios\n\n", app.Manifest.URL, len(scenarios))

	var results []Result
	for _, sc := range scenarios {
		failures, err := runScenario(ctx, opts, app, sc)
		if err != nil {
			return nil, err
		}
		results = append(results, Result{Scenario: sc, Failures: failures})
		if len(failures) == 0 {
			fmt.Fprintf(out, "PASS %s: %s\n", sc.File, sc.Name)
		} else {
			fmt.Fprintf(out, "FAIL %s: %s\n", sc.File, sc.Name)
			for _, f := range failures {
				for _, line := range strings.Split(strings.TrimRight(f, "\n"), "\n") {
					fmt.Fprintf(out, "  %s\n", line)
				}
			}
		}
	}

	passed, failed := 0, 0
	for _, r := range results {
		if len(r.Failures) == 0 {
			passed++
		} else {
			failed++
		}
	}
	fmt.Fprintf(out, "\n%d passed, %d failed, %d total\n", passed, failed, len(results))
	return results, nil
}

// runScenario executes one scenario per HARNESS.md § Execution semantics.
// The returned error is fatal for the whole run; scenario-level problems
// (including infrastructure trouble driving this scenario) become failures.
func runScenario(ctx context.Context, opts Options, app *harness.App, sc scenario.Scenario) ([]string, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	// Seed always, including the empty model (HARNESS.md step 2).
	seedItems, err := seedItems(sc.Given)
	if err != nil {
		return []string{err.Error()}, nil
	}
	if err := app.Seed(ctx, seedItems); err != nil {
		return []string{err.Error()}, nil
	}

	// Fresh browser context (step 1: nothing survives from earlier scenarios).
	session, err := browser.NewSession(ctx, opts.ChromePath)
	if err != nil {
		return nil, err // browser trouble is fatal for the run
	}
	defer session.Close()

	// Step 3: full page load of url + route.
	url := strings.TrimRight(app.Manifest.URL, "/") + sc.Route
	if err := session.Navigate(url); err != nil {
		return []string{fmt.Sprintf("navigating to %s: %v", url, err)}, nil
	}

	// Step 4: WHEN actions in order.
	for _, a := range sc.When {
		if err := session.Do(a, app.Manifest.URL); err != nil {
			return []string{err.Error()}, nil
		}
	}

	// Step 5: THEN sections. All of them are evaluated even after a failure,
	// so a run reports every broken expectation at once. Each assertion is
	// polled up to AssertWait: rendering and persistence happen after the
	// last input event, and the docs give no synchronization signal.
	var failures []string

	if sc.HasThenPage {
		if f := assertPage(session, sc.ThenPage, opts.AssertWait); f != "" {
			failures = append(failures, f)
		}
	}
	if sc.HasThenModel {
		if f := assertModel(ctx, app, sc.ThenModel, opts.AssertWait); f != "" {
			failures = append(failures, f)
		}
	}
	for _, c := range sc.Checks {
		if f := assertCheck(session, c, opts.AssertWait); f != "" {
			failures = append(failures, f)
		}
	}

	// Always-on invariants: non-empty persisted ids, and displayed data-id
	// consistent with the persisted item of the same title.
	persisted, err := app.Read(ctx)
	if err != nil {
		failures = append(failures, err.Error())
	} else {
		displayed, err := session.DisplayedItems()
		if err != nil {
			failures = append(failures, fmt.Sprintf("reading displayed todos: %v", err))
		} else {
			failures = append(failures, VerifyIntegrity(persisted, displayed)...)
		}
	}
	return failures, nil
}

// seedItems turns the GIVEN model into the seed payload. Explicit ids are
// passed verbatim; omitted ids get a generated unique opaque id, since the
// seed command stores ids verbatim and has no notion of an absent id.
func seedItems(m scenario.Model) ([]harness.Item, error) {
	used := map[string]bool{}
	for _, t := range m {
		if t.ID != "" {
			used[t.ID] = true
		}
	}
	items := make([]harness.Item, 0, len(m))
	for _, t := range m {
		id := t.ID
		if id == "" {
			for id == "" || used[id] {
				var buf [4]byte
				if _, err := rand.Read(buf[:]); err != nil {
					return nil, err
				}
				id = "gen-" + hex.EncodeToString(buf[:])
			}
			used[id] = true
		}
		items = append(items, harness.Item{ID: id, Title: t.Title, Completed: t.Completed})
	}
	return items, nil
}

func assertPage(session *browser.Session, expected []string, wait time.Duration) string {
	var actual []string
	var lastErr error
	deadline := time.Now().Add(wait)
	for {
		raw, err := session.Project()
		lastErr = err
		if err == nil {
			actual = EraseUnrequestedIDs(raw, expected)
			if linesEqual(actual, expected) {
				return ""
			}
		}
		if time.Now().After(deadline) {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	if lastErr != nil {
		return fmt.Sprintf("THEN page: projecting the page: %v", lastErr)
	}
	return "THEN page: projection mismatch\n" +
		SideBySide(expected, actual, "expected", "actual")
}

func assertModel(ctx context.Context, app *harness.App, expected scenario.Model, wait time.Duration) string {
	var problems []string
	var actual []harness.Item
	var lastErr error
	deadline := time.Now().Add(wait)
	for {
		items, err := app.Read(ctx)
		lastErr = err
		if err == nil {
			problems = CompareModel(expected, items)
			if len(problems) == 0 {
				return ""
			}
			actual = items
		}
		if time.Now().After(deadline) {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	if lastErr != nil {
		return fmt.Sprintf("THEN model: %v", lastErr)
	}
	return fmt.Sprintf("THEN model: mismatch: %s\n%s",
		strings.Join(problems, "; "),
		SideBySide(expected.Notation(), ItemsToModel(actual).Notation(), "expected", "actual"))
}

func assertCheck(session *browser.Session, c scenario.Check, wait time.Duration) string {
	var detail string
	var lastErr error
	deadline := time.Now().Add(wait)
	for {
		ok, d, err := session.Check(c)
		lastErr = err
		if err == nil {
			if ok {
				return ""
			}
			detail = d
		}
		if time.Now().After(deadline) {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	if lastErr != nil {
		return fmt.Sprintf("THEN check %q: %v", c.Raw, lastErr)
	}
	return fmt.Sprintf("THEN check %q failed: %s", c.Raw, detail)
}
