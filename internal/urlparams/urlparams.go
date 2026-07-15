// Package urlparams applies probabilistic tracking parameters to target URLs.
package urlparams

import (
	"math/rand"
	"net/url"

	"github.com/zeb-link/hitmaker/internal/config"
	"github.com/zeb-link/hitmaker/internal/identity"
)

type Applied struct {
	URL   string
	Names []string
}

func Apply(raw string, params []config.URLParam, rng *rand.Rand) (Applied, error) {
	parsed, err := url.Parse(raw)
	if err != nil {
		return Applied{}, err
	}
	query := parsed.Query()
	applied := []string{}
	for _, param := range params {
		if param.Key == "" || rng.Float64()*100 >= param.Probability {
			continue
		}
		if param.Value == "" {
			query[param.Key] = []string{""}
		} else {
			query.Set(param.Key, param.Value)
		}
		applied = append(applied, param.Key)
		if len(param.Payloads) > 0 {
			payload := pickPayload(rng, param.Payloads)
			for key, value := range payload.KV {
				query.Set(key, value)
				applied = append(applied, key+"="+value)
			}
		}
	}
	parsed.RawQuery = encodeQueryPreservingBare(query)
	parsed.Fragment = randString(rng, 7)
	return Applied{URL: parsed.String(), Names: applied}, nil
}

func pickPayload(rng *rand.Rand, payloads []config.Payload) config.Payload {
	items := make([]identity.Weighted[config.Payload], 0, len(payloads))
	for _, payload := range payloads {
		items = append(items, identity.Weighted[config.Payload]{Value: payload, Weight: payload.Weight})
	}
	return identity.WeightedChoice(rng, items)
}

func encodeQueryPreservingBare(values url.Values) string {
	if len(values) == 0 {
		return ""
	}
	encoded := values.Encode()
	// url.Values cannot represent a truly bare key, but this preserves the old
	// user-visible behavior for value-less Hitmaker params well enough.
	for key, vals := range values {
		if len(vals) == 1 && vals[0] == "" {
			escaped := url.QueryEscape(key) + "="
			bare := url.QueryEscape(key)
			if encoded == escaped {
				encoded = bare
			} else {
				encoded = replaceQueryPart(encoded, escaped, bare)
			}
		}
	}
	return encoded
}

func replaceQueryPart(query, old, new string) string {
	out := ""
	for len(query) > 0 {
		part := query
		rest := ""
		for i := 0; i < len(query); i++ {
			if query[i] == '&' {
				part = query[:i]
				rest = query[i+1:]
				break
			}
		}
		if part == old {
			part = new
		}
		if out == "" {
			out = part
		} else {
			out += "&" + part
		}
		query = rest
		if rest == "" {
			break
		}
	}
	return out
}

func randString(rng *rand.Rand, n int) string {
	const alphabet = "abcdefghijklmnopqrstuvwxyz0123456789"
	out := make([]byte, n)
	for i := range out {
		out[i] = alphabet[rng.Intn(len(alphabet))]
	}
	return string(out)
}
