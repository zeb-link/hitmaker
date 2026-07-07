package identity

import (
	"math/rand"
	"net"
	"strings"
	"testing"
)

func TestWeightedChoice(t *testing.T) {
	rng := rand.New(rand.NewSource(4))
	items := []Weighted[string]{{Value: "never", Weight: 0}, {Value: "always", Weight: 10}}
	for i := 0; i < 20; i++ {
		if got := WeightedChoice(rng, items); got != "always" {
			t.Fatalf("got %q, want always", got)
		}
	}
}

func TestFakeIPIsValidAndBounded(t *testing.T) {
	g := New(1, 8)
	for i := 0; i < 200; i++ {
		ident := g.Next(60, 0, 1)
		if parsed := net.ParseIP(ident.IP); parsed == nil {
			t.Fatalf("invalid ip %q", ident.IP)
		}
	}
	if got := g.SubnetCount(); got > 8 {
		t.Fatalf("subnet ring grew to %d, want <= 8", got)
	}
}

func TestReturningVisitorReusesSubnet(t *testing.T) {
	g := New(2, 8)
	first := g.Next(60, 0, 1).IP
	second := g.Next(60, 0, 0).IP
	if subnet(first) != subnet(second) {
		t.Fatalf("subnet %s was not reused by %s", subnet(first), second)
	}
}

func subnet(ip string) string {
	parts := strings.Split(ip, ".")
	return strings.Join(parts[:3], ".")
}
