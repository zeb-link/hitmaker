// Package identity generates per-request browser and geo identity.
package identity

import (
	"fmt"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

type Weighted[T any] struct {
	Value  T
	Weight float64
}

type Location struct {
	Country   string
	City      string
	Region    string
	Latitude  string
	Longitude string
}

type RequestIdentity struct {
	UserAgent      string
	AcceptLanguage string
	Referer        string
	Location       Location
	IP             string
}

type Generator struct {
	mu      sync.Mutex
	rng     *rand.Rand
	subnets *subnetRing
	botPool []Weighted[string]
}

func New(seed int64, subnetLimit int) *Generator {
	if seed == 0 {
		seed = time.Now().UnixNano()
	}
	return &Generator{
		rng:     rand.New(rand.NewSource(seed)),
		subnets: newSubnetRing(subnetLimit),
		botPool: BotPool(BotFilter{}),
	}
}

// UseBots restricts the "unknown/bot" pool to the agents matching filter.
// An empty filter (the default) draws from the whole catalog.
func (g *Generator) UseBots(filter BotFilter) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.botPool = BotPool(filter)
}

func (g *Generator) Next(deviceRatio, unknownRatio int, uniqueIPProb float64) RequestIdentity {
	g.mu.Lock()
	defer g.mu.Unlock()

	var ua string
	if g.rng.Float64()*100 < float64(unknownRatio) {
		ua = WeightedChoice(g.rng, g.botPool)
	} else if g.rng.Float64()*100 < float64(deviceRatio) {
		ua = choice(g.rng, DesktopUserAgents)
	} else {
		ua = choice(g.rng, MobileUserAgents)
	}
	loc := choice(g.rng, Locations)
	return RequestIdentity{
		UserAgent:      ua,
		AcceptLanguage: choice(g.rng, AcceptLanguages),
		Referer:        choice(g.rng, Referers),
		Location:       loc,
		IP:             g.fakeIP(loc.Country, uniqueIPProb),
	}
}

func (g *Generator) SubnetCount() int {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.subnets.Len()
}

func (g *Generator) fakeIP(country string, uniqueIPProb float64) string {
	if g.rng.Float64() < uniqueIPProb || g.subnets.Len() == 0 {
		firstChoices := IPFirstOctets[country]
		if len(firstChoices) == 0 {
			firstChoices = IPFirstOctets["US"]
		}
		o1 := choice(g.rng, firstChoices)
		o2 := g.rng.Intn(256)
		o3 := g.rng.Intn(256)
		o4 := 1 + g.rng.Intn(254)
		g.subnets.Add(fmt.Sprintf("%d.%d.%d", o1, o2, o3))
		return fmt.Sprintf("%d.%d.%d.%d", o1, o2, o3, o4)
	}
	subnet := g.subnets.Random(g.rng)
	return fmt.Sprintf("%s.%d", subnet, 1+g.rng.Intn(254))
}

func ApplyBrowserHeaders(headers http.Header, ident RequestIdentity) {
	headers.Set("User-Agent", ident.UserAgent)
	headers.Set("Accept-Language", ident.AcceptLanguage)
	headers.Set("Referer", ident.Referer)
	headers.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
}

func ApplyVercelGeoHeaders(headers http.Header, ident RequestIdentity) {
	headers.Set("x-forwarded-for", ident.IP)
	headers.Set("x-real-ip", ident.IP)
	headers.Set("x-vercel-ip-country", ident.Location.Country)
	headers.Set("x-vercel-ip-city", ident.Location.City)
	headers.Set("x-vercel-ip-country-region", ident.Location.Region)
	headers.Set("x-vercel-ip-latitude", ident.Location.Latitude)
	headers.Set("x-vercel-ip-longitude", ident.Location.Longitude)
}

func WeightedChoice[T any](rng *rand.Rand, items []Weighted[T]) T {
	if len(items) == 0 {
		var zero T
		return zero
	}
	total := 0.0
	for _, item := range items {
		if item.Weight > 0 {
			total += item.Weight
		}
	}
	if total <= 0 {
		return items[len(items)-1].Value
	}
	roll := rng.Float64() * total
	for _, item := range items {
		if item.Weight <= 0 {
			continue
		}
		roll -= item.Weight
		if roll <= 0 {
			return item.Value
		}
	}
	return items[len(items)-1].Value
}

func choice[T any](rng *rand.Rand, items []T) T {
	return items[rng.Intn(len(items))]
}

type subnetRing struct {
	limit int
	order []string
	seen  map[string]struct{}
}

func newSubnetRing(limit int) *subnetRing {
	if limit <= 0 {
		limit = 2048
	}
	return &subnetRing{limit: limit, seen: map[string]struct{}{}}
}

func (r *subnetRing) Add(subnet string) {
	if _, ok := r.seen[subnet]; ok {
		return
	}
	if len(r.order) >= r.limit {
		old := r.order[0]
		delete(r.seen, old)
		copy(r.order, r.order[1:])
		r.order[len(r.order)-1] = subnet
		r.seen[subnet] = struct{}{}
		return
	}
	r.order = append(r.order, subnet)
	r.seen[subnet] = struct{}{}
}

func (r *subnetRing) Random(rng *rand.Rand) string {
	return r.order[rng.Intn(len(r.order))]
}

func (r *subnetRing) Len() int {
	return len(r.order)
}
