// Package tui provides a transport-free V1 TUI client scaffold.
package tui

import (
	"context"

	"data-toolkit/internal/application"
	"data-toolkit/internal/config"
)

type Client struct {
	api application.API
}

func New(api application.API) Client {
	return Client{api: api}
}

func (client Client) Capabilities(ctx context.Context) (application.Capabilities, error) {
	return client.api.Capabilities(ctx)
}

func (client Client) ValidateConfig(ctx context.Context, payload []byte, format config.Format) (application.ConfigValidation, error) {
	return client.api.ValidateConfig(ctx, payload, format)
}
