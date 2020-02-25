package platform

import (
	"go-alrd/secret"
	"net/http"
)

type AuthTransport struct {
	rt http.RoundTripper
}

func (at *AuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// add api key
	req.Header.Set("X-ELS-APIKey", secret.APIKey)
	return at.rt.RoundTrip(req)
}

func NewAuthTransport(rt http.RoundTripper) *AuthTransport {
	if rt == nil {
		rt = http.DefaultTransport
	}
	return &AuthTransport{rt}
}
