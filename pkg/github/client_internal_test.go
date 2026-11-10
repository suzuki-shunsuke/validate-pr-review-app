package github

import (
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"net/http"
	"testing"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/golang-jwt/jwt/v4"
)

type errSigner struct{}

func (errSigner) Sign(_ jwt.Claims) (string, error) {
	return "", errors.New("access denied")
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

// doRequest sends a request through ghinstallation's transports and http.Client, as the actual client does,
// and returns the error.
func doRequest(t *testing.T, base http.RoundTripper, signer Signer) error {
	t.Helper()
	atr, err := ghinstallation.NewAppsTransportWithOptions(base, 1, ghinstallation.WithSigner(signer))
	if err != nil {
		t.Fatal(err)
	}
	client := &http.Client{Transport: ghinstallation.NewFromAppsTransport(atr, 2)}
	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "https://api.github.com/repos/foo/bar", nil)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := client.Do(req)
	if err == nil {
		resp.Body.Close()
		t.Fatal("an error is expected")
	}
	return err //nolint:wrapcheck
}

func TestIsSignError(t *testing.T) {
	t.Parallel()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	rsaSigner := ghinstallation.NewRSASigner(jwt.SigningMethodRS256, key)
	networkErr := roundTripperFunc(func(*http.Request) (*http.Response, error) {
		return nil, errors.New("connection refused")
	})
	tests := []struct {
		name   string
		base   http.RoundTripper
		signer Signer
		want   bool
	}{
		{
			name:   "signing fails",
			base:   networkErr,
			signer: errSigner{},
			want:   true,
		},
		{
			name:   "getting an installation access token fails due to a network error",
			base:   networkErr,
			signer: rsaSigner,
			want:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := doRequest(t, tt.base, tt.signer)
			if got := isSignError(err); got != tt.want {
				t.Errorf("isSignError(%v) = %v, want %v", err, got, tt.want)
			}
		})
	}
	if isSignError(nil) {
		t.Error("isSignError(nil) = true, want false")
	}
}
