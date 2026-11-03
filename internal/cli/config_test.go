package cli

import (
	"testing"

	"github.com/zeb-link/hitmaker/v2/internal/config"
)

func TestSetProviderURLKeySelectsThatProvider(t *testing.T) {
	for _, provider := range config.ProxyProviders {
		for _, key := range []string{provider.Name + "url", provider.Name + "_url", "origin." + provider.Name + "url"} {
			cfg := config.Default()
			if err := setKey(&cfg, key, provider.Example); err != nil {
				t.Fatalf("setKey(%s): %v", key, err)
			}
			if cfg.Origin.ProviderName() != provider.Name || cfg.Origin.ProviderURL(provider.Name) != provider.Example {
				t.Fatalf("setKey(%s) -> provider %q url %q", key, cfg.Origin.ProviderName(), cfg.Origin.ProviderURL(provider.Name))
			}
		}
	}
}

func TestSetProxyURLTargetsActiveProviderOnly(t *testing.T) {
	first, second := config.ProxyProviders[0], config.ProxyProviders[1]
	cfg := config.Default()
	cfg.Origin.SetProviderSetting(first.Name, "url", first.Example)
	if err := setKey(&cfg, "provider", second.Name); err != nil {
		t.Fatal(err)
	}
	if err := setKey(&cfg, "proxy-url", second.Example); err != nil {
		t.Fatal(err)
	}
	if cfg.Origin.ProviderURL(second.Name) != second.Example || cfg.Origin.ProviderURL(first.Name) != first.Example {
		t.Fatalf("providers = %v", cfg.Origin.Providers)
	}
}

func TestSetUnknownProviderFails(t *testing.T) {
	cfg := config.Default()
	if err := setKey(&cfg, "provider", "nope"); err == nil {
		t.Fatal("expected unknown provider error")
	}
}
