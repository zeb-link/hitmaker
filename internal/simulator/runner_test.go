package simulator

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/kerns/hitmaker/internal/config"
)

func testConfig() config.Config {
	c := config.Default()
	c.Requests.URLParams = nil // keep probes deterministic
	return c
}

// Guards the reported bug: a redirect tester must report the redirect's own
// status by default and not chase the destination.
func TestNoFollowReportsRedirectStatus(t *testing.T) {
	final := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	defer final.Close()
	redir := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, final.URL, http.StatusFound)
	}))
	defer redir.Close()

	r1, err := New(context.Background(), Options{Config: testConfig(), Targets: []string{redir.URL}})
	if err != nil {
		t.Fatal(err)
	}
	if got := r1.Probe(redir.URL); got.Status != 302 {
		t.Fatalf("no-follow status = %d, want 302 (err=%q)", got.Status, got.Err)
	}

	r2, err := New(context.Background(), Options{Config: testConfig(), Targets: []string{redir.URL}, Follow: true})
	if err != nil {
		t.Fatal(err)
	}
	if got := r2.Probe(redir.URL); got.Status != 200 {
		t.Fatalf("follow status = %d, want 200 (err=%q)", got.Status, got.Err)
	}
}

func TestSeedReproducibleIdentity(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) }))
	defer srv.Close()
	r1, _ := New(context.Background(), Options{Config: testConfig(), Targets: []string{srv.URL}, Seed: 7})
	r2, _ := New(context.Background(), Options{Config: testConfig(), Targets: []string{srv.URL}, Seed: 7})
	a := r1.Probe(srv.URL)
	b := r2.Probe(srv.URL)
	if a.UserAgent != b.UserAgent || a.IP != b.IP {
		t.Fatalf("seed not reproducible: %q/%q vs %q/%q", a.UserAgent, a.IP, b.UserAgent, b.IP)
	}
}

func TestAutoModeUsesVercelHeadersForLocalTargets(t *testing.T) {
	seen := make(chan http.Header, 1)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen <- r.Header.Clone()
		w.WriteHeader(200)
	}))
	defer srv.Close()

	cfg := testConfig()
	cfg.Origin.Mode = config.ModeAuto
	r, err := New(context.Background(), Options{Config: cfg, Targets: []string{srv.URL}})
	if err != nil {
		t.Fatal(err)
	}
	got := r.Probe(srv.URL)
	if got.Err != "" || got.Status != 200 {
		t.Fatalf("probe status=%d err=%q", got.Status, got.Err)
	}
	headers := <-seen
	if headers.Get("x-vercel-ip-country") == "" || headers.Get("x-forwarded-for") == "" {
		t.Fatalf("auto local target did not receive Vercel geo headers: %#v", headers)
	}
}

func TestAutoModeRequiresProxyProviderForPublicTargets(t *testing.T) {
	cfg := testConfig()
	cfg.Origin.Mode = config.ModeAuto
	_, err := New(context.Background(), Options{Config: cfg, Targets: []string{"https://example.com/a"}})
	if err == nil {
		t.Fatal("expected missing proxy provider error")
	}
	if !strings.Contains(err.Error(), "iproyal provider requires") {
		t.Fatalf("error = %q, want missing iproyal provider", err)
	}
}

// Guards the reported hang: shutdown must return within the grace window.
func TestStopAndWaitReturnsPromptly(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) }))
	defer srv.Close()
	r, _ := New(context.Background(), Options{Config: testConfig(), Targets: []string{srv.URL}})
	r.Start()
	done := make(chan struct{})
	go func() { r.StopAndWait(2 * time.Second); close(done) }()
	select {
	case <-done:
	case <-time.After(4 * time.Second):
		t.Fatal("StopAndWait did not return within the grace window")
	}
}
