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

package acp

import (
	"context"
	"errors"
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/traefik/hub-agent-kubernetes/pkg/acp/basicauth"
	"github.com/traefik/hub-agent-kubernetes/pkg/acp/jwt"
	"github.com/traefik/hub-agent-kubernetes/pkg/acp/oidc"
	hubv1alpha1 "github.com/traefik/hub-agent-kubernetes/pkg/crd/api/hub/v1alpha1"
	kerror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
)

// Config is the configuration of an Access Control Policy. It is used to setup ACP handlers.
type Config struct {
	JWT       *jwt.Config
	BasicAuth *basicauth.Config
	OIDC      *oidc.Config
}

// ConfigFromPolicy returns an ACP configuration for the given policy.
func ConfigFromPolicy(policy *hubv1alpha1.AccessControlPolicy, kubeClientset *clientset.Clientset) *Config {
	switch {
	case policy.Spec.JWT != nil:
		jwtCfg := policy.Spec.JWT

		return &Config{
			JWT: &jwt.Config{
				SigningSecret:              jwtCfg.SigningSecret,
				SigningSecretBase64Encoded: jwtCfg.SigningSecretBase64Encoded,
				PublicKey:                  jwtCfg.PublicKey,
				JWKsFile:                   jwt.FileOrContent(jwtCfg.JWKsFile),
				JWKsURL:                    jwtCfg.JWKsURL,
				StripAuthorizationHeader:   jwtCfg.StripAuthorizationHeader,
				ForwardHeaders:             jwtCfg.ForwardHeaders,
				TokenQueryKey:              jwtCfg.TokenQueryKey,
				Claims:                     jwtCfg.Claims,
			},
		}

	case policy.Spec.BasicAuth != nil:
		basicCfg := policy.Spec.BasicAuth

		return &Config{
			BasicAuth: &basicauth.Config{
				Users:                    basicCfg.Users,
				Realm:                    basicCfg.Realm,
				StripAuthorizationHeader: basicCfg.StripAuthorizationHeader,
				ForwardUsernameHeader:    basicCfg.ForwardUsernameHeader,
			},
		}

	case policy.Spec.OIDC != nil:
		oidcCfg := policy.Spec.OIDC

		conf := &Config{
			OIDC: &oidc.Config{
				ClientSecret:   oidcCfg.ClientSecret,
				Issuer:         oidcCfg.Issuer,
				ClientID:       oidcCfg.ClientID,
				RedirectURL:    oidcCfg.RedirectURL,
				LogoutURL:      oidcCfg.LogoutURL,
				Scopes:         oidcCfg.Scopes,
				AuthParams:     oidcCfg.AuthParams,
				ForwardHeaders: oidcCfg.ForwardHeaders,
				Claims:         oidcCfg.Claims,
			},
		}

		if oidcCfg.Secret != nil {
			conf.OIDC.Secret = &oidc.SecretReference{
				Name:      oidcCfg.Secret.Name,
				Namespace: oidcCfg.Secret.Namespace,
			}
		}

		if oidcCfg.StateCookie != nil {
			conf.OIDC.StateCookie = &oidc.AuthStateCookie{
				Secret:   oidcCfg.StateCookie.Secret,
				Path:     oidcCfg.StateCookie.Path,
				Domain:   oidcCfg.StateCookie.Domain,
				SameSite: oidcCfg.StateCookie.SameSite,
				Secure:   oidcCfg.StateCookie.Secure,
			}
		}

		if oidcCfg.StateCookie != nil {
			conf.OIDC.Session = &oidc.AuthSession{
				Secret:   oidcCfg.Session.Secret,
				Path:     oidcCfg.Session.Path,
				Domain:   oidcCfg.Session.Domain,
				SameSite: oidcCfg.Session.SameSite,
				Secure:   oidcCfg.Session.Secure,
				Refresh:  oidcCfg.Session.Refresh,
			}
		}

		if oidcCfg.TLS != nil {
			conf.OIDC.TLS = &oidc.TLS{
				CABundle:           oidcCfg.TLS.CABundle,
				InsecureSkipVerify: oidcCfg.TLS.InsecureSkipVerify,
			}
		}

		var oidcSecret oidcSecret
		if oidcCfg.Secret != nil && oidcCfg.Secret.Name != "" && kubeClientset != nil {
			var err error
			oidcSecret, err = getOIDCSecret(oidcCfg.Secret.Name, oidcCfg.Secret.Namespace, kubeClientset)
			if err != nil {
				log.Error().Err(err).Msg("getOIDCSecret")
				return &Config{}
			}
			conf.OIDC.ClientSecret = oidcSecret.ClientSecret

			if conf.OIDC.StateCookie == nil {
				conf.OIDC.StateCookie = &oidc.AuthStateCookie{}
			}
			conf.OIDC.StateCookie.Secret = oidcSecret.StateCookieKey

			if conf.OIDC.Session == nil {
				conf.OIDC.Session = &oidc.AuthSession{}
			}
			conf.OIDC.Session.Secret = oidcSecret.StateCookieKey
		}

		return conf
	default:
		return &Config{}
	}
}

func getOIDCSecret(secretName, namespace string, kubeClientset *clientset.Clientset) (oidcSecret, error) {
	if namespace == "" {
		namespace = "default"
	}

	secret, err := kubeClientset.CoreV1().Secrets(namespace).Get(context.Background(), secretName, metav1.GetOptions{})
	if err != nil && !kerror.IsNotFound(err) {
		return oidcSecret{}, fmt.Errorf("get secret: %w", err)
	}

	clientSecret, ok := secret.Data["clientSecret"]
	if !ok {
		return oidcSecret{}, errors.New("missing client secret")
	}

	sessionKey, ok := secret.Data["sessionKey"]
	if !ok {
		return oidcSecret{}, errors.New("missing session key")
	}

	stateCookieKey, ok := secret.Data["stateCookieKey"]
	if !ok {
		return oidcSecret{}, errors.New("missing state cookie key")
	}

	return oidcSecret{
		ClientSecret:   string(clientSecret),
		SessionKey:     string(sessionKey),
		StateCookieKey: string(stateCookieKey),
	}, nil
}

type oidcSecret struct {
	ClientSecret   string
	SessionKey     string
	StateCookieKey string
}
