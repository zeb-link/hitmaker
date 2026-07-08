package proxy

import "testing"

func TestIsPublicDomainHost(t *testing.T) {
	tests := []struct {
		host string
		want bool
	}{
		{host: "example.com", want: true},
		{host: "www.gov.uk", want: true},
		{host: "foo.pages.dev", want: true},
		{host: "localhost", want: false},
		{host: "app.local", want: false},
		{host: "api.internal", want: false},
		{host: "service.lan", want: false},
		{host: "example.test", want: false},
		{host: "127.0.0.1", want: false},
		{host: "::1", want: false},
		{host: "printer", want: false},
		{host: "example.notarealtld", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.host, func(t *testing.T) {
			if got := IsPublicDomainHost(tt.host); got != tt.want {
				t.Fatalf("IsPublicDomainHost(%q) = %v, want %v", tt.host, got, tt.want)
			}
		})
	}
}
