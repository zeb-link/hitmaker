package config

import (
	"fmt"
	"strings"
)

// ProxyProvider describes a paid proxy service Hitmaker can route through.
//
// This list is the single source for everything user-facing about providers:
// the `provider` config value, the <NAME>_URL environment variable, the
// `config set <name>url` key, the TUI picker, and help text. The adapter that
// actually dials each provider lives in internal/proxy, and a test there fails
// if the two lists drift apart.
type ProxyProvider struct {
	// Name is the config value, e.g. "iproyal".
	Name string
	// Label is the display name, e.g. "IPRoyal".
	Label string
	// Example is the endpoint URL shape shown in help text.
	Example string
}

// ProxyProviders lists the supported proxy services. The first entry is the
// default when no provider is configured.
var ProxyProviders = []ProxyProvider{
	{Name: "iproyal", Label: "IPRoyal", Example: "http://user:pass@geo.iproyal.com:12321"},
	{Name: "evomi", Label: "Evomi", Example: "http://user:pass@core-residential.evomi.com:1000"},
}

// DefaultProxyProvider is the provider used when none is configured.
var DefaultProxyProvider = ProxyProviders[0].Name

// legacyProxyProvider owns credentials in pre-2.6 config files that don't
// name a provider: IPRoyal was the only one then. Fixed on purpose, so a
// change of default can't reassign old credentials.
const legacyProxyProvider = "iproyal"

// EnvVar is the environment variable that supplies this provider's endpoint
// URL, e.g. IPROYAL_URL.
func (p ProxyProvider) EnvVar() string {
	return strings.ToUpper(p.Name) + "_URL"
}

// LookupProxyProvider finds a supported provider by name (case-insensitive).
func LookupProxyProvider(name string) (ProxyProvider, bool) {
	name = strings.ToLower(strings.TrimSpace(name))
	for _, provider := range ProxyProviders {
		if provider.Name == name {
			return provider, true
		}
	}
	return ProxyProvider{}, false
}

// ProxyProviderNames returns the supported provider names in list order.
func ProxyProviderNames() []string {
	names := make([]string, len(ProxyProviders))
	for i, provider := range ProxyProviders {
		names[i] = provider.Name
	}
	return names
}

func unknownProviderError(name string) error {
	return fmt.Errorf("unknown proxy provider %q (supported: %s)", name, strings.Join(ProxyProviderNames(), ", "))
}

// ProviderName resolves the active proxy provider, falling back to the default
// when none is set.
func (o OriginConfig) ProviderName() string {
	if o.Provider == "" {
		return DefaultProxyProvider
	}
	return o.Provider
}

// ActiveProviderSettings returns the settings of the active proxy provider.
// The map may be nil; callers only read from it.
func (o OriginConfig) ActiveProviderSettings() map[string]string {
	return o.Providers[o.ProviderName()]
}

// ProviderURL returns the endpoint URL stored for the named provider.
func (o OriginConfig) ProviderURL(name string) string {
	return o.Providers[name]["url"]
}

// SetProviderSetting stores one setting for the named provider without
// touching any other provider's settings.
func (o *OriginConfig) SetProviderSetting(name, key, value string) {
	if o.Providers == nil {
		o.Providers = map[string]map[string]string{}
	}
	if o.Providers[name] == nil {
		o.Providers[name] = map[string]string{}
	}
	o.Providers[name][key] = value
}
