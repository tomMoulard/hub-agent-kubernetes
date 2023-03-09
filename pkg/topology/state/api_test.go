/*
Copyright (C) 2022-2023 Traefik Labs

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

package state

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	hubv1alpha1 "github.com/traefik/hub-agent-kubernetes/pkg/crd/api/hub/v1alpha1"
	hubkubemock "github.com/traefik/hub-agent-kubernetes/pkg/crd/generated/client/hub/clientset/versioned/fake"
	traefikkubemock "github.com/traefik/hub-agent-kubernetes/pkg/crd/generated/client/traefik/clientset/versioned/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubemock "k8s.io/client-go/kubernetes/fake"
)

func TestFetcher_GetAPIs(t *testing.T) {
	want := map[string]*API{
		"api@api-ns": {
			Name:       "api",
			Namespace:  "api-ns",
			Labels:     map[string]string{"key": "value"},
			PathPrefix: "/api",
			Service: APIService{
				Name: "api-service",
				Port: APIServiceBackendPort{
					Number: 80,
				},
				OpenAPISpec: OpenAPISpec{
					URL: "https://example.com/api.json",
				},
			},
		},
	}

	objects := loadK8sObjects(t, "fixtures/api/api.yml")

	kubeClient := kubemock.NewSimpleClientset()
	// Faking having Hub CRDs installed on cluster.
	kubeClient.Resources = append(kubeClient.Resources, &metav1.APIResourceList{
		GroupVersion: hubv1alpha1.SchemeGroupVersion.String(),
		APIResources: []metav1.APIResource{
			{Kind: "APIPortal"},
		},
	})
	traefikClient := traefikkubemock.NewSimpleClientset()
	hubClient := hubkubemock.NewSimpleClientset(objects...)

	f, err := watchAll(context.Background(), kubeClient, traefikClient, hubClient, "v1.20.1")
	require.NoError(t, err)

	got, err := f.getAPIs()
	require.NoError(t, err)

	assert.Equal(t, want, got)
}
