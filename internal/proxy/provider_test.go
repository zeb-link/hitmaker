package proxy

import (
	"context"
	"net/http"
	"sort"
	"strings"
	"testing"

	"github.com/zeb-link/hitmaker/v2/internal/config"
)

// The user-facing provider list (config.ProxyProviders) and the adapters
// registered here must name the same providers.
func TestRegistryMatchesConfigProviders(t *testing.T) {
	registry := DefaultRegistry()
	var registered []string
	for name := range registry.providers {
		registered = append(registered, name)
	}
	configured := config.ProxyProviderNames()
	sort.Strings(registered)
	sort.Strings(configured)
	if strings.Join(registered, ",") != strings.Join(configured, ",") {
		t.Fatalf("registry %v != config.ProxyProviders %v", registered, configured)
	}
}

func TestEveryProviderBuildsRotatingTransportFromItsExample(t *testing.T) {
	registry := DefaultRegistry()
	for _, entry := range config.ProxyProviders {
		t.Run(entry.Name, func(t *testing.T) {
			provider, _ := registry.Provider(entry.Name)
			cfg := ProviderConfig{"url": entry.Example}
			if err := provider.Validate(cfg); err != nil {
				t.Fatalf("Validate(example) = %v", err)
			}
			transport, err := provider.Transport(context.Background(), cfg)
			if err != nil {
				t.Fatal(err)
			}
			if !transport.DisableKeepAlives {
				t.Fatal("rotating gateway must not reuse connections")
			}
			req, _ := http.NewRequest(http.MethodGet, "https://example.com", nil)
			proxyURL, err := transport.Proxy(req)
			if err != nil || proxyURL == nil || proxyURL.String() != entry.Example {
				t.Fatalf("proxy = %v (err %v), want %s", proxyURL, err, entry.Example)
			}
		})
	}
}

func TestMissingURLNamesTheProviderEnvVar(t *testing.T) {
	registry := DefaultRegistry()
	for _, entry := range config.ProxyProviders {
		provider, _ := registry.Provider(entry.Name)
		err := provider.Validate(ProviderConfig{})
		if err == nil || !strings.Contains(err.Error(), entry.EnvVar()) {
			t.Fatalf("%s: error %v should name %s", entry.Name, err, entry.EnvVar())
		}
	}
}

func TestBuildTransportUsesSelectedProvider(t *testing.T) {
	origin := config.OriginConfig{Mode: config.ModeProxy}
	for _, entry := range config.ProxyProviders {
		origin.SetProviderSetting(entry.Name, "url", entry.Example)
	}
	for _, entry := range config.ProxyProviders {
		origin.Provider = entry.Name
		transport, err := DefaultRegistry().BuildTransport(context.Background(), origin)
		if err != nil {
			t.Fatal(err)
		}
		req, _ := http.NewRequest(http.MethodGet, "https://example.com", nil)
		proxyURL, _ := transport.(*http.Transport).Proxy(req)
		if proxyURL.String() != entry.Example {
			t.Fatalf("provider %s routed via %s", entry.Name, proxyURL)
		}
	}
}
