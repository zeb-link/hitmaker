package urlparams

import (
	"math/rand"
	"strings"
	"testing"

	"github.com/zeb-link/hitmaker/v2/internal/config"
)

func TestApplyProbabilityAndFragmentCacheBust(t *testing.T) {
	got, err := Apply("https://example.com/a?x=1", []config.URLParam{
		{Key: "qr", Value: "1", Probability: 100},
		{Key: "skip", Value: "1", Probability: 0},
	}, rand.New(rand.NewSource(1)))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got.URL, "x=1") || !strings.Contains(got.URL, "qr=1") {
		t.Fatalf("missing query params: %s", got.URL)
	}
	if strings.Contains(got.URL, "skip=1") {
		t.Fatalf("probability 0 param was applied: %s", got.URL)
	}
	if !strings.Contains(got.URL, "#") {
		t.Fatalf("cache bust fragment missing: %s", got.URL)
	}
}

func TestApplyBareParam(t *testing.T) {
	got, err := Apply("https://example.com/a", []config.URLParam{
		{Key: "qr", Probability: 100},
	}, rand.New(rand.NewSource(2)))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got.URL, "?qr#") {
		t.Fatalf("bare param not preserved: %s", got.URL)
	}
}

func TestApplyWeightedPayload(t *testing.T) {
	got, err := Apply("https://example.com/a", []config.URLParam{{
		Key: "qr", Value: "1", Probability: 100,
		Payloads: []config.Payload{
			{Name: "zero", Weight: 0, KV: map[string]string{"city": "never"}},
			{Name: "hit", Weight: 1, KV: map[string]string{"city": "copenhagen"}},
		},
	}}, rand.New(rand.NewSource(3)))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got.URL, "city=copenhagen") {
		t.Fatalf("payload missing: %s", got.URL)
	}
}
