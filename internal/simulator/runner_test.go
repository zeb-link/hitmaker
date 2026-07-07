package simulator

import (
	"context"
	"net/http"
	"net/http/httptest"
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
