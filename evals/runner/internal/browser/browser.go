// Package browser drives a real Chrome over CDP (via chromedp) and implements
// the page projection, action vocabulary, and check registry of evals/DSL.md
// against the DOM vocabulary of spec/main-screen-template.html.
package browser

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/chromedp/chromedp"
)

// Session is one fresh, isolated browser: its own Chrome process with its own
// temporary user data dir, so no client-side state can survive from any
// previous scenario (HARNESS.md execution semantics, step 1).
type Session struct {
	Ctx     context.Context
	cancels []context.CancelFunc
	tmpDir  string
}

// NewSession launches a fresh headless Chrome. chromePath overrides the
// executable lookup when non-empty.
func NewSession(parent context.Context, chromePath string) (*Session, error) {
	tmp, err := os.MkdirTemp("", "evals-chrome-*")
	if err != nil {
		return nil, err
	}
	opts := append([]chromedp.ExecAllocatorOption{},
		chromedp.DefaultExecAllocatorOptions[:]...)
	opts = append(opts, chromedp.UserDataDir(tmp))
	if chromePath != "" {
		opts = append(opts, chromedp.ExecPath(chromePath))
	}
	allocCtx, cancelAlloc := chromedp.NewExecAllocator(parent, opts...)
	ctx, cancelCtx := chromedp.NewContext(allocCtx)
	s := &Session{
		Ctx:     ctx,
		cancels: []context.CancelFunc{cancelCtx, cancelAlloc},
		tmpDir:  tmp,
	}
	// Start the browser now so failures surface here, not mid-scenario.
	if err := chromedp.Run(ctx); err != nil {
		s.Close()
		return nil, fmt.Errorf("launching Chrome: %w", err)
	}
	return s, nil
}

// Close tears the browser down and removes its temporary profile.
func (s *Session) Close() {
	for _, c := range s.cancels {
		c()
	}
	if s.tmpDir != "" {
		os.RemoveAll(s.tmpDir)
	}
}

// run executes chromedp actions with a per-operation timeout.
func (s *Session) run(timeout time.Duration, actions ...chromedp.Action) error {
	ctx, cancel := context.WithTimeout(s.Ctx, timeout)
	defer cancel()
	return chromedp.Run(ctx, actions...)
}

// Navigate performs a full page load of the given URL.
func (s *Session) Navigate(url string) error {
	return s.run(20*time.Second, chromedp.Navigate(url))
}

// eval evaluates a JS expression and stores the JSON result in out.
func (s *Session) eval(expr string, out interface{}) error {
	return s.run(10*time.Second, chromedp.Evaluate(expr, out))
}
