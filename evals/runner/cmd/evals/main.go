// Command evals runs the eval scenario suite against an app implementation,
// per evals/DSL.md and evals/HARNESS.md.
//
// Usage:
//
//	go run ./cmd/evals -app <path-to-app-dir> [-scenarios <dir>] [-chrome <path>]
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"time"

	"todomvc/evals/runner/internal/run"
)

func main() {
	appDir := flag.String("app", "", "app directory containing harness.json (required)")
	scenariosDir := flag.String("scenarios", "", "scenarios directory (default: the evals/scenarios directory found near the runner)")
	chromePath := flag.String("chrome", "", "Chrome executable (default: found automatically)")
	wait := flag.Duration("wait", 3*time.Second, "how long THEN assertions are polled before failing")
	flag.Parse()

	if *appDir == "" {
		fmt.Fprintln(os.Stderr, "usage: evals -app <path-to-app-dir> [flags]")
		flag.PrintDefaults()
		os.Exit(2)
	}

	dir := *scenariosDir
	if dir == "" {
		var err error
		dir, err = findScenariosDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "evals: %v (use -scenarios to point at it)\n", err)
			os.Exit(2)
		}
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	results, err := run.Run(ctx, run.Options{
		AppDir:       *appDir,
		ScenariosDir: dir,
		ChromePath:   *chromePath,
		AssertWait:   *wait,
		Out:          os.Stdout,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "evals: %v\n", err)
		os.Exit(2)
	}
	for _, r := range results {
		if len(r.Failures) > 0 {
			os.Exit(1)
		}
	}
}

// findScenariosDir locates evals/scenarios relative to the working
// directory: it tries the directory itself, then evals/scenarios and
// scenarios beneath it, walking upward — so it works from the repo root,
// from evals/, and from evals/runner/.
func findScenariosDir() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for i := 0; i < 6; i++ {
		for _, cand := range []string{
			filepath.Join(dir, "evals", "scenarios"),
			filepath.Join(dir, "scenarios"),
		} {
			if st, err := os.Stat(cand); err == nil && st.IsDir() {
				return cand, nil
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", fmt.Errorf("could not find an evals/scenarios directory near %s", dir)
}
