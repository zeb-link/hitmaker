// Package proxy contains paid-provider routing adapters.
package proxy

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/kerns/hitmaker/internal/config"
)

type ProviderConfig map[string]string

type Provider interface {
	Name() string
	Validate(cfg ProviderConfig) error
	Transport(ctx context.Context, cfg ProviderConfig) (*http.Transport, error)
}

type Registry struct {
	providers map[string]Provider
}

func DefaultRegistry() Registry {
	r := Registry{providers: map[string]Provider{}}
	r.Register(IPRoyal{})
	return r
}

func (r Registry) Register(provider Provider) {
	r.providers[provider.Name()] = provider
}

func (r Registry) Provider(name string) (Provider, bool) {
	provider, ok := r.providers[name]
	return provider, ok
}

func (r Registry) BuildTransport(ctx context.Context, origin config.OriginConfig) (*http.Transport, error) {
	if origin.Mode != config.ModeProxy {
		return BaseTransport(), nil
	}
	providerName := origin.Provider
	if providerName == "" {
		providerName = "iproyal"
	}
	provider, ok := r.Provider(providerName)
	if !ok {
		return nil, fmt.Errorf("unknown proxy provider %q", providerName)
	}
	cfg := ProviderConfig(origin.ProviderConfig)
	if err := provider.Validate(cfg); err != nil {
		return nil, err
	}
	return provider.Transport(ctx, cfg)
}

func BaseTransport() *http.Transport {
	dialer := &cachedDialer{
		inner: &net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 30 * time.Second,
		},
		ttl: 60 * time.Second,
	}
	return &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           dialer.DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          128,
		MaxIdleConnsPerHost:   32,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
}

// rotatingProxyTransport is BaseTransport with connection reuse turned OFF. Use
// it from any provider whose gateway assigns a fresh exit IP PER NEW TCP
// CONNECTION (rotating residential/mobile endpoints — IPRoyal's geo gateway,
// most Bright Data / Oxylabs rotating ports, etc.). With keep-alive on, the
// shared worker client pins each connection to the exit it first dialed, so a
// whole run collapses onto a handful of sticky IPs and the geo looks stuck to a
// couple of countries. Disabling keep-alive forces a fresh dial per request, so
// every request draws a new exit and traffic distributes globally as intended.
// Sticky-session providers should NOT use this — they want the connection held.
func rotatingProxyTransport() *http.Transport {
	transport := BaseTransport()
	transport.DisableKeepAlives = true
	transport.MaxIdleConnsPerHost = -1
	return transport
}

type IPRoyal struct{}

func (IPRoyal) Name() string { return "iproyal" }

func (IPRoyal) Validate(cfg ProviderConfig) error {
	raw := cfg["url"]
	if raw == "" {
		return errors.New("iproyal provider requires providerConfig.url or IPROYAL_URL")
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("iproyal url is invalid: %w", err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" && parsed.Scheme != "socks5" {
		return errors.New("iproyal url must use http, https, or socks5")
	}
	if parsed.Host == "" {
		return errors.New("iproyal url must include a host")
	}
	return nil
}

func (IPRoyal) Transport(ctx context.Context, cfg ProviderConfig) (*http.Transport, error) {
	_ = ctx
	raw := cfg["url"]
	parsed, err := url.Parse(raw)
	if err != nil {
		return nil, err
	}
	// IPRoyal's geo gateway rotates the exit IP per new connection, so we must
	// not pool/reuse connections or the run sticks to a few exits (see
	// rotatingProxyTransport).
	transport := rotatingProxyTransport()
	transport.Proxy = http.ProxyURL(parsed)
	return transport, nil
}

type cachedDialer struct {
	inner *net.Dialer
	ttl   time.Duration
	mu    sync.Mutex
	cache map[string]dnsEntry
}

type dnsEntry struct {
	addrs  []net.IPAddr
	expiry time.Time
}

func (d *cachedDialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return d.inner.DialContext(ctx, network, address)
	}
	addrs, err := d.lookup(ctx, host)
	if err != nil || len(addrs) == 0 {
		return d.inner.DialContext(ctx, network, address)
	}
	var last error
	for _, addr := range addrs {
		conn, err := d.inner.DialContext(ctx, network, net.JoinHostPort(addr.IP.String(), port))
		if err == nil {
			return conn, nil
		}
		last = err
	}
	return nil, last
}

func (d *cachedDialer) lookup(ctx context.Context, host string) ([]net.IPAddr, error) {
	if net.ParseIP(host) != nil {
		return []net.IPAddr{{IP: net.ParseIP(host)}}, nil
	}
	now := time.Now()
	d.mu.Lock()
	if d.cache == nil {
		d.cache = map[string]dnsEntry{}
	}
	if entry, ok := d.cache[host]; ok && now.Before(entry.expiry) {
		addrs := append([]net.IPAddr(nil), entry.addrs...)
		d.mu.Unlock()
		return addrs, nil
	}
	d.mu.Unlock()

	resolver := net.DefaultResolver
	addrs, err := resolver.LookupIPAddr(ctx, host)
	if err != nil {
		return nil, err
	}
	d.mu.Lock()
	d.cache[host] = dnsEntry{addrs: append([]net.IPAddr(nil), addrs...), expiry: now.Add(d.ttl)}
	d.mu.Unlock()
	return addrs, nil
}
