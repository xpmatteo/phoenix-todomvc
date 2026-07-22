// Package harness implements the adapter contract defined in evals/HARNESS.md:
// the harness.json manifest, the start/seed/read commands, and app lifecycle.
package harness

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

// Item is the JSON shape exchanged with the seed and read commands.
type Item struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Completed bool   `json:"completed"`
}

// Manifest is app/harness.json.
type Manifest struct {
	Start string `json:"start"`
	URL   string `json:"url"`
	Seed  string `json:"seed"`
	Read  string `json:"read"`
}

// App is a loaded manifest bound to its app directory.
type App struct {
	Dir      string
	Manifest Manifest

	cmd      *exec.Cmd
	startLog bytes.Buffer
}

// Load reads <appDir>/harness.json.
func Load(appDir string) (*App, error) {
	abs, err := filepath.Abs(appDir)
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(filepath.Join(abs, "harness.json"))
	if err != nil {
		return nil, fmt.Errorf("reading manifest: %w", err)
	}
	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parsing harness.json: %w", err)
	}
	for k, v := range map[string]string{"start": m.Start, "url": m.URL, "seed": m.Seed, "read": m.Read} {
		if strings.TrimSpace(v) == "" {
			return nil, fmt.Errorf("harness.json: missing %q", k)
		}
	}
	return &App{Dir: abs, Manifest: m}, nil
}

// StartTimeout is how long Start polls the app URL for an HTTP 200 (HARNESS.md).
const StartTimeout = 60 * time.Second

// Start launches the start command in its own process group and polls the
// manifest URL until it responds with HTTP 200 or StartTimeout elapses.
func (a *App) Start(ctx context.Context) error {
	cmd := exec.Command("/bin/sh", "-c", a.Manifest.Start)
	cmd.Dir = a.Dir
	cmd.Stdout = &a.startLog
	cmd.Stderr = &a.startLog
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("starting app: %w", err)
	}
	a.cmd = cmd

	deadline := time.Now().Add(StartTimeout)
	client := &http.Client{Timeout: 2 * time.Second}
	for {
		if ctx.Err() != nil {
			a.Stop()
			return ctx.Err()
		}
		if cmd.ProcessState != nil {
			return fmt.Errorf("app exited before becoming ready:\n%s", a.startLog.String())
		}
		resp, err := client.Get(a.Manifest.URL)
		if err == nil {
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return nil
			}
		}
		if time.Now().After(deadline) {
			a.Stop()
			return fmt.Errorf("app at %s not ready within %s; app output:\n%s",
				a.Manifest.URL, StartTimeout, a.startLog.String())
		}
		time.Sleep(200 * time.Millisecond)
	}
}

// Stop kills the app's whole process group.
func (a *App) Stop() {
	if a.cmd == nil || a.cmd.Process == nil {
		return
	}
	pgid := a.cmd.Process.Pid
	syscall.Kill(-pgid, syscall.SIGTERM)
	done := make(chan struct{})
	go func() {
		a.cmd.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		syscall.Kill(-pgid, syscall.SIGKILL)
		<-done
	}
	a.cmd = nil
}

// StartLog returns the accumulated stdout+stderr of the start command.
func (a *App) StartLog() string { return a.startLog.String() }

// Seed runs the seed command, feeding the model as JSON on stdin. The model
// replaces the entire persisted state; an empty slice seeds the empty array.
func (a *App) Seed(ctx context.Context, model []Item) error {
	if model == nil {
		model = []Item{}
	}
	payload, err := json.Marshal(model)
	if err != nil {
		return err
	}
	cmd := exec.CommandContext(ctx, "/bin/sh", "-c", a.Manifest.Seed)
	cmd.Dir = a.Dir
	cmd.Stdin = bytes.NewReader(payload)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("seed command failed: %w\n%s", err, out)
	}
	return nil
}

// Read runs the read command and parses its stdout as the persisted model.
func (a *App) Read(ctx context.Context) ([]Item, error) {
	cmd := exec.CommandContext(ctx, "/bin/sh", "-c", a.Manifest.Read)
	cmd.Dir = a.Dir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("read command failed: %w\n%s", err, stderr.String())
	}
	var items []Item
	dec := json.NewDecoder(bytes.NewReader(stdout.Bytes()))
	if err := dec.Decode(&items); err != nil {
		return nil, fmt.Errorf("read command output is not a JSON model array: %w\noutput: %s",
			err, stdout.String())
	}
	return items, nil
}
