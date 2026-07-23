package adapter

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"data-toolkit/internal/config"

	"golang.org/x/oauth2"
)

// GoogleOAuthTokenSourceResolver obtains a refresh-capable token source from
// an external credential store. Implementations must keep tokens and client
// secrets outside workflows and ordinary Data Toolkit configuration.
type GoogleOAuthTokenSourceResolver func(context.Context, config.CredentialRequest) (oauth2.TokenSource, error)

// ExternalGoogleOAuthProvider adapts an external token store to the
// credential-provider boundary used by Google adapters. It never owns or
// serializes credential material.
type ExternalGoogleOAuthProvider struct {
	resolve GoogleOAuthTokenSourceResolver
}

// NewExternalGoogleOAuthProvider creates a provider backed by an external
// refresh-capable token source resolver.
func NewExternalGoogleOAuthProvider(resolve GoogleOAuthTokenSourceResolver) (*ExternalGoogleOAuthProvider, error) {
	if resolve == nil {
		return nil, errors.New("external Google OAuth token source resolver is required")
	}
	return &ExternalGoogleOAuthProvider{resolve: resolve}, nil
}

// GoogleHTTPClient returns an authorized client while keeping the token source
// opaque to adapters. Errors from the external store and token refresh are
// deliberately replaced with safe errors because their text may contain
// credential material.
func (provider *ExternalGoogleOAuthProvider) GoogleHTTPClient(ctx context.Context, request config.CredentialRequest) (*http.Client, error) {
	if provider == nil || provider.resolve == nil {
		return nil, errors.New("external Google OAuth provider is not initialized")
	}
	if ctx == nil {
		return nil, errors.New("Google OAuth context is required")
	}
	normalized, err := normalizeGoogleCredentialRequest(request)
	if err != nil {
		return nil, err
	}

	source, err := provider.resolve(ctx, normalized)
	if err != nil || source == nil {
		return nil, fmt.Errorf("external Google OAuth credentials are unavailable for profile %q", normalized.Profile)
	}

	return oauth2.NewClient(ctx, safeGoogleTokenSource{source: source}), nil
}

func normalizeGoogleCredentialRequest(request config.CredentialRequest) (config.CredentialRequest, error) {
	request.Profile = strings.TrimSpace(request.Profile)
	if request.Profile == "" {
		return config.CredentialRequest{}, errors.New("Google OAuth profile is required")
	}
	if len(request.Scopes) == 0 {
		return config.CredentialRequest{}, errors.New("at least one Google OAuth scope is required")
	}

	seen := make(map[string]struct{}, len(request.Scopes))
	scopes := make([]string, 0, len(request.Scopes))
	for _, scope := range request.Scopes {
		if !allowedGoogleOAuthScope(scope) {
			return config.CredentialRequest{}, fmt.Errorf("Google OAuth scope %q is not permitted", scope)
		}
		if _, duplicate := seen[scope]; duplicate {
			continue
		}
		seen[scope] = struct{}{}
		scopes = append(scopes, scope)
	}
	request.Scopes = scopes
	return request, nil
}

type safeGoogleTokenSource struct {
	source oauth2.TokenSource
}

func (source safeGoogleTokenSource) Token() (*oauth2.Token, error) {
	token, err := source.source.Token()
	if err != nil || token == nil {
		return nil, errors.New("external Google OAuth token is unavailable")
	}
	return token, nil
}

var _ config.CredentialProvider = (*ExternalGoogleOAuthProvider)(nil)
