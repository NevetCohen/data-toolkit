package adapter

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"data-toolkit/internal/config"

	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

// GoogleAPIClients contains Sheets and Drive clients authorized for one
// narrowly defined adapter use case.
type GoogleAPIClients struct {
	Sheets *sheets.Service
	Drive  *drive.Service
}

const (
	googleSheetsReadOnlyScope    = sheets.SpreadsheetsReadonlyScope
	googleSheetsWriteScope       = sheets.SpreadsheetsScope
	googleDriveMetadataReadScope = drive.DriveMetadataReadonlyScope
	googleDriveFileWriteScope    = drive.DriveFileScope
)

// NewGoogleReadOnlyClients authorizes displayed-value reads and spreadsheet
// discovery without granting Sheets writes or full Drive content access.
func NewGoogleReadOnlyClients(ctx context.Context, provider config.CredentialProvider, profile string) (GoogleAPIClients, error) {
	return newGoogleAPIClients(ctx, provider, profile, []string{
		googleSheetsReadOnlyScope,
		googleDriveMetadataReadScope,
	})
}

// NewGoogleWriteClients authorizes Sheets writes and only the Drive files the
// application creates or the user explicitly opens with it. It does not grant
// unrestricted Drive access.
func NewGoogleWriteClients(ctx context.Context, provider config.CredentialProvider, profile string) (GoogleAPIClients, error) {
	return newGoogleAPIClients(ctx, provider, profile, []string{
		googleSheetsWriteScope,
		googleDriveFileWriteScope,
	})
}

func newGoogleAPIClients(ctx context.Context, provider config.CredentialProvider, profile string, scopes []string) (GoogleAPIClients, error) {
	if ctx == nil {
		return GoogleAPIClients{}, errors.New("Google API client context is required")
	}
	if provider == nil {
		return GoogleAPIClients{}, errors.New("Google credential provider is required")
	}
	profile = strings.TrimSpace(profile)
	if profile == "" {
		return GoogleAPIClients{}, errors.New("Google OAuth profile is required")
	}

	requestedScopes := append([]string(nil), scopes...)
	client, err := provider.GoogleHTTPClient(ctx, config.CredentialRequest{
		Profile: profile,
		Scopes:  requestedScopes,
	})
	if err != nil {
		// External providers are not trusted to return secret-free error text.
		return GoogleAPIClients{}, fmt.Errorf("authorize Google APIs for profile %q", profile)
	}
	if client == nil {
		return GoogleAPIClients{}, fmt.Errorf("authorize Google APIs for profile %q: provider returned no client", profile)
	}

	sheetsClient, err := sheets.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return GoogleAPIClients{}, fmt.Errorf("construct Sheets API client: %w", err)
	}
	driveClient, err := drive.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return GoogleAPIClients{}, fmt.Errorf("construct Drive API client: %w", err)
	}
	return GoogleAPIClients{Sheets: sheetsClient, Drive: driveClient}, nil
}

func allowedGoogleOAuthScope(scope string) bool {
	switch scope {
	case googleSheetsReadOnlyScope, googleSheetsWriteScope, googleDriveMetadataReadScope, googleDriveFileWriteScope:
		return true
	default:
		return false
	}
}
