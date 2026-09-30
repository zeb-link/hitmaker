package identity

import (
	"math/rand"
	"net"
	"net/http"
	"strings"
	"testing"
)

func TestEveryLocationHasARegionName(t *testing.T) {
	for _, loc := range Locations {
		if _, ok := RegionNames[loc.Country+"-"+loc.Region]; !ok {
			t.Errorf("%s/%s: region %s-%s has no RegionNames entry", loc.Country, loc.City, loc.Country, loc.Region)
		}
	}
}

func TestCloudflareHeadersCarryRegionNameAndCode(t *testing.T) {
	headers := make(http.Header)
	loc := Location{Country: "US", City: "Austin", Region: "TX"}
	ApplyCloudflareGeoHeaders(headers, RequestIdentity{Location: loc})
	if got := headers.Get("x-hitmaker-region"); got != "Texas" {
		t.Errorf("x-hitmaker-region = %q, want Texas", got)
	}
	if got := headers.Get("x-hitmaker-region-code"); got != "TX" {
		t.Errorf("x-hitmaker-region-code = %q, want TX", got)
	}
}

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

func TestEveryLocationCountryHasOctets(t *testing.T) {
	for _, loc := range Locations {
		if len(IPFirstOctets[loc.Country]) == 0 {
			t.Fatalf("location %s/%s has no IP octets — IPs would fall back to US and mismatch the geo", loc.City, loc.Country)
		}
	}
}
