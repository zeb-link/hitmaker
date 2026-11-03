package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// clearProviderEnv unsets every provider env var for the test, so a developer
// shell that exports IPROYAL_URL can't leak into the result.
func clearProviderEnv(t *testing.T) {
	t.Helper()
	keys := []string{"HITMAKER_PROVIDER", "HITMAKER_MODE", "PROXY_MODE", "PROXY_SERVICE_URL"}
	for _, provider := range ProxyProviders {
		keys = append(keys, provider.EnvVar())
	}
	for _, key := range keys {
		if old, ok := os.LookupEnv(key); ok {
			t.Cleanup(func() { _ = os.Setenv(key, old) })
			_ = os.Unsetenv(key)
		}
	}
}

func TestProviderNameDefaultsToFirstProvider(t *testing.T) {
	if got := Default().Origin.ProviderName(); got != ProxyProviders[0].Name {
		t.Fatalf("ProviderName() = %q, want %q", got, ProxyProviders[0].Name)
	}
}

func TestValidateRejectsUnknownProvider(t *testing.T) {
	cfg := Default()
	cfg.Origin.Provider = "nope"
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected unknown provider error")
	}
	for _, provider := range ProxyProviders {
		cfg.Origin.Provider = provider.Name
		if err := cfg.Validate(); err != nil {
			t.Fatalf("provider %q rejected: %v", provider.Name, err)
		}
	}
}

func TestSwitchingProviderKeepsOtherCredentials(t *testing.T) {
	cfg := Default()
	for _, provider := range ProxyProviders {
		cfg.Origin.SetProviderSetting(provider.Name, "url", provider.Example)
	}
	for _, provider := range ProxyProviders {
		cfg.Origin.Provider = provider.Name
		if got := cfg.Origin.ActiveProviderSettings()["url"]; got != provider.Example {
			t.Fatalf("%s url = %q, want %q", provider.Name, got, provider.Example)
		}
	}
}

func TestLegacyProviderConfigFileMigrates(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	legacy := map[string]any{
		"origin": map[string]any{
			"mode":           "proxy",
			"provider":       "iproyal",
			"providerConfig": map[string]string{"url": "http://u:p@geo.iproyal.com:12321"},
		},
	}
	data, _ := json.Marshal(legacy)
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	partial, ok, err := readPartial(path)
	if err != nil || !ok {
		t.Fatalf("readPartial: ok=%v err=%v", ok, err)
	}
	cfg := Default()
	merge(&cfg, partial)
	if got := cfg.Origin.ProviderURL("iproyal"); got != "http://u:p@geo.iproyal.com:12321" {
		t.Fatalf("migrated iproyal url = %q", got)
	}
	out, _ := json.Marshal(cfg)
	var round map[string]map[string]any
	_ = json.Unmarshal(out, &round)
	if _, stale := round["origin"]["providerConfig"]; stale {
		t.Fatalf("saved config still writes providerConfig: %s", out)
	}
}

func TestLegacyProviderConfigStaysWithItsFileWhenAnotherLayerSwitches(t *testing.T) {
	other := ProxyProviders[1]
	legacyURL := "http://u:p@geo.iproyal.com:12321"
	path := filepath.Join(t.TempDir(), ".hitmaker.json")
	data, _ := json.Marshal(map[string]any{
		"origin": map[string]any{"providerConfig": map[string]string{"url": legacyURL}},
	})
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	partial, _, err := readPartial(path)
	if err != nil {
		t.Fatal(err)
	}
	cfg := Default()
	cfg.Origin.Provider = other.Name
	cfg.Origin.SetProviderSetting(other.Name, "url", other.Example)
	merge(&cfg, partial)
	if got := cfg.Origin.ProviderURL(other.Name); got != other.Example {
		t.Fatalf("legacy credentials leaked into %s: %q", other.Name, got)
	}
	if got := cfg.Origin.ProviderURL(legacyProxyProvider); got != legacyURL {
		t.Fatalf("legacy %s url = %q", legacyProxyProvider, got)
	}
}

func TestProviderEnvSelectsProviderAndKeepsSavedCredentials(t *testing.T) {
	clearProviderEnv(t)
	saved, other := ProxyProviders[0], ProxyProviders[1]
	cfg := Default()
	cfg.Origin.Provider = saved.Name
	cfg.Origin.SetProviderSetting(saved.Name, "url", saved.Example)

	t.Setenv(other.EnvVar(), other.Example)
	env, err := envPartial()
	if err != nil {
		t.Fatal(err)
	}
	merge(&cfg, env)
	if cfg.Origin.ProviderName() != other.Name {
		t.Fatalf("provider = %q, want %q (its env var is the only one set)", cfg.Origin.ProviderName(), other.Name)
	}
	if cfg.Origin.Mode != ModeProxy {
		t.Fatalf("mode = %q, want proxy", cfg.Origin.Mode)
	}
	if got := cfg.Origin.ProviderURL(saved.Name); got != saved.Example {
		t.Fatalf("saved %s url lost: %q", saved.Name, got)
	}
}

func TestSeveralProviderEnvsLeaveSelectionAlone(t *testing.T) {
	clearProviderEnv(t)
	cfg := Default()
	cfg.Origin.Provider = ProxyProviders[1].Name
	for _, provider := range ProxyProviders {
		t.Setenv(provider.EnvVar(), provider.Example)
	}
	env, err := envPartial()
	if err != nil {
		t.Fatal(err)
	}
	merge(&cfg, env)
	if cfg.Origin.ProviderName() != ProxyProviders[1].Name {
		t.Fatalf("provider = %q, want saved %q", cfg.Origin.ProviderName(), ProxyProviders[1].Name)
	}

	t.Setenv("HITMAKER_PROVIDER", ProxyProviders[0].Name)
	env, err = envPartial()
	if err != nil {
		t.Fatal(err)
	}
	merge(&cfg, env)
	if cfg.Origin.ProviderName() != ProxyProviders[0].Name {
		t.Fatalf("HITMAKER_PROVIDER ignored: provider = %q", cfg.Origin.ProviderName())
	}
}

func TestRedactedMasksEveryProvider(t *testing.T) {
	cfg := Default()
	for _, provider := range ProxyProviders {
		cfg.Origin.SetProviderSetting(provider.Name, "url", provider.Example)
	}
	redacted := cfg.Redacted()
	for _, provider := range ProxyProviders {
		if got := redacted.Origin.ProviderURL(provider.Name); got != "********" {
			t.Fatalf("%s url not redacted: %q", provider.Name, got)
		}
		if cfg.Origin.ProviderURL(provider.Name) != provider.Example {
			t.Fatalf("Redacted mutated the original %s url", provider.Name)
		}
	}
}
