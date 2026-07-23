package adapter

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"data-toolkit/internal/config"

	"golang.org/x/oauth2"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/sheets/v4"
)

func TestExternalGoogleOAuthProviderUsesExternalTokenSource(t *testing.T) {
	const accessToken = "oauth-access-secret"
	var captured config.CredentialRequest
	provider, err := NewExternalGoogleOAuthProvider(func(_ context.Context, request config.CredentialRequest) (oauth2.TokenSource, error) {
		captured = request
		return oauth2.StaticTokenSource(&oauth2.Token{AccessToken: accessToken, TokenType: "Bearer"}), nil
	})
	if err != nil {
		t.Fatal(err)
	}

	client, err := provider.GoogleHTTPClient(context.Background(), config.CredentialRequest{
		Profile: "  personal  ",
		Scopes: []string{
			sheets.SpreadsheetsReadonlyScope,
			sheets.SpreadsheetsReadonlyScope,
			drive.DriveMetadataReadonlyScope,
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if got := request.Header.Get("Authorization"); got != "Bearer "+accessToken {
			t.Errorf("Authorization = %q", got)
		}
		response.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	response, err := client.Get(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	if err := response.Body.Close(); err != nil {
		t.Fatal(err)
	}

	wantScopes := []string{sheets.SpreadsheetsReadonlyScope, drive.DriveMetadataReadonlyScope}
	if captured.Profile != "personal" {
		t.Fatalf("profile = %q", captured.Profile)
	}
	if !reflect.DeepEqual(captured.Scopes, wantScopes) {
		t.Fatalf("scopes = %v, want %v", captured.Scopes, wantScopes)
	}
}

func TestExternalGoogleOAuthProviderRejectsBroadScopes(t *testing.T) {
	called := false
	provider, err := NewExternalGoogleOAuthProvider(func(context.Context, config.CredentialRequest) (oauth2.TokenSource, error) {
		called = true
		return nil, nil
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = provider.GoogleHTTPClient(context.Background(), config.CredentialRequest{
		Profile: "personal",
		Scopes:  []string{drive.DriveScope},
	})
	if err == nil {
		t.Fatal("broad Drive scope was accepted")
	}
	if called {
		t.Fatal("external token source was called for a forbidden scope")
	}
}

func TestExternalGoogleOAuthProviderSanitizesResolverAndRefreshErrors(t *testing.T) {
	const secret = "refresh-token-secret"
	provider, err := NewExternalGoogleOAuthProvider(func(context.Context, config.CredentialRequest) (oauth2.TokenSource, error) {
		return nil, errors.New("load failed for " + secret)
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = provider.GoogleHTTPClient(context.Background(), config.CredentialRequest{
		Profile: "personal",
		Scopes:  []string{sheets.SpreadsheetsReadonlyScope},
	})
	if err == nil || strings.Contains(err.Error(), secret) {
		t.Fatalf("resolver error was not sanitized: %v", err)
	}

	provider, err = NewExternalGoogleOAuthProvider(func(context.Context, config.CredentialRequest) (oauth2.TokenSource, error) {
		return failingTokenSource{err: errors.New("refresh failed for " + secret)}, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	client, err := provider.GoogleHTTPClient(context.Background(), config.CredentialRequest{
		Profile: "personal",
		Scopes:  []string{sheets.SpreadsheetsReadonlyScope},
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.Get("https://example.invalid")
	if err == nil || strings.Contains(err.Error(), secret) {
		t.Fatalf("refresh error was not sanitized: %v", err)
	}
}

func TestGoogleClientConstructionUsesLeastPrivilegeScopes(t *testing.T) {
	provider := &capturingCredentialProvider{
		client: &http.Client{},
	}

	readClients, err := NewGoogleReadOnlyClients(context.Background(), provider, "personal")
	if err != nil {
		t.Fatal(err)
	}
	if readClients.Sheets == nil || readClients.Drive == nil {
		t.Fatal("read-only Google clients were not constructed")
	}
	if len(provider.requests) != 1 {
		t.Fatalf("credential requests after read client = %d, want 1", len(provider.requests))
	}
	wantRead := []string{sheets.SpreadsheetsReadonlyScope, drive.DriveMetadataReadonlyScope}
	if !reflect.DeepEqual(provider.requests[0].Scopes, wantRead) {
		t.Fatalf("read scopes = %v, want %v", provider.requests[0].Scopes, wantRead)
	}

	writeClients, err := NewGoogleWriteClients(context.Background(), provider, "personal")
	if err != nil {
		t.Fatal(err)
	}
	if writeClients.Sheets == nil || writeClients.Drive == nil {
		t.Fatal("write Google clients were not constructed")
	}
	if len(provider.requests) != 2 {
		t.Fatalf("credential requests = %d, want 2", len(provider.requests))
	}
	wantWrite := []string{sheets.SpreadsheetsScope, drive.DriveFileScope}
	if !reflect.DeepEqual(provider.requests[1].Scopes, wantWrite) {
		t.Fatalf("write scopes = %v, want %v", provider.requests[1].Scopes, wantWrite)
	}
	for _, request := range provider.requests {
		for _, scope := range request.Scopes {
			if scope == drive.DriveScope || scope == drive.DriveReadonlyScope {
				t.Fatalf("broad Drive scope requested: %q", scope)
			}
		}
	}
}

func TestGoogleClientConstructionSanitizesProviderErrors(t *testing.T) {
	const secret = "client-secret-value"
	provider := &capturingCredentialProvider{err: errors.New("provider failed with " + secret)}
	_, err := NewGoogleReadOnlyClients(context.Background(), provider, "personal")
	if err == nil || strings.Contains(err.Error(), secret) {
		t.Fatalf("provider error was not sanitized: %v", err)
	}
}

type failingTokenSource struct {
	err error
}

func (source failingTokenSource) Token() (*oauth2.Token, error) {
	return nil, source.err
}

type capturingCredentialProvider struct {
	requests []config.CredentialRequest
	client   *http.Client
	err      error
}

func (provider *capturingCredentialProvider) GoogleHTTPClient(_ context.Context, request config.CredentialRequest) (*http.Client, error) {
	request.Scopes = append([]string(nil), request.Scopes...)
	provider.requests = append(provider.requests, request)
	return provider.client, provider.err
}
