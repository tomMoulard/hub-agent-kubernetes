/*
Copyright (C) 2022 Traefik Labs

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published
by the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.
*/

package oidc

import (
	"context"
	"errors"
	"fmt"

	"github.com/coreos/go-oidc/v3/oidc"
)

// Config holds the configuration for the OIDC middleware.
type Config struct {
	Issuer       string
	ClientID     string
	ClientSecret string
	SecretName   string
	TLS          *TLS

	RedirectURL string
	LogoutURL   string
	Scopes      []string
	AuthParams  map[string]string
	StateCookie *AuthStateCookie
	Session     *AuthSession

	// ForwardHeaders defines headers that should be added to the request and populated with values extracted from the ID token.
	ForwardHeaders map[string]string
	// Claims defines an expression to perform validation on the ID token. For example:
	//     Equals(`grp`, `admin`) && Equals(`scope`, `deploy`)
	Claims string
}

// AuthStateCookie carries the state cookie configuration.
type AuthStateCookie struct {
	Secret   string
	Path     string
	Domain   string
	SameSite string
	Secure   bool
}

// AuthSession carries session and session cookie configuration.
type AuthSession struct {
	Secret   string
	Path     string
	Domain   string
	SameSite string
	Secure   bool
	Refresh  *bool
}

// ApplyDefaultValues applies default values on the given dynamic configuration.
func ApplyDefaultValues(cfg *Config) {
	if len(cfg.Scopes) == 0 {
		cfg.Scopes = []string{"openid"}
	}

	if cfg.StateCookie == nil {
		cfg.StateCookie = &AuthStateCookie{}
	}

	if cfg.StateCookie.Path == "" {
		cfg.StateCookie.Path = "/"
	}

	if cfg.StateCookie.SameSite == "" {
		cfg.StateCookie.SameSite = "lax"
	}

	if cfg.Session == nil {
		cfg.Session = &AuthSession{}
	}

	if cfg.Session.Path == "" {
		cfg.Session.Path = "/"
	}

	if cfg.Session.SameSite == "" {
		cfg.Session.SameSite = "lax"
	}

	if cfg.Session.Refresh == nil {
		cfg.Session.Refresh = ptrBool(true)
	}

	if cfg.RedirectURL == "" {
		cfg.RedirectURL = "/callback"
	}
}

// Validate validates configuration.
func (cfg *Config) Validate() error {
	ApplyDefaultValues(cfg)

	if cfg.Issuer == "" {
		return errors.New("missing issuer")
	}

	if cfg.ClientID == "" {
		return errors.New("missing client ID")
	}

	if cfg.ClientSecret == "" {
		return errors.New("missing client secret")
	}

	if cfg.Session.Secret == "" {
		return errors.New("missing session secret")
	}

	switch len(cfg.Session.Secret) {
	case 16, 24, 32:
		break
	default:
		return errors.New("session secret must be 16, 24 or 32 characters long")
	}

	if cfg.StateCookie.Secret == "" {
		return errors.New("missing state secret")
	}

	switch len(cfg.StateCookie.Secret) {
	case 16, 24, 32:
		break
	default:
		return errors.New("state secret must be 16, 24 or 32 characters long")
	}

	if cfg.RedirectURL == "" {
		return errors.New("missing redirect URL")
	}

	return nil
}

// ptrBool returns a pointer to boolean.
func ptrBool(v bool) *bool {
	return &v
}

// BuildProvider returns a provider instance from given auth source.
func BuildProvider(ctx context.Context, cfg *Config) (*oidc.Provider, error) {
	provider, err := oidc.NewProvider(ctx, cfg.Issuer)
	if err != nil {
		return nil, fmt.Errorf("unable to create provider: %w", err)
	}

	return provider, nil
}
