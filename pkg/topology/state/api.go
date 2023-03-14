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
	"k8s.io/apimachinery/pkg/labels"
)

func (f *Fetcher) getAPIs() (map[string]*API, error) {
	apis, err := f.hub.Hub().V1alpha1().APIs().Lister().List(labels.Everything())
	if err != nil {
		return nil, err
	}

	result := make(map[string]*API)
	for _, api := range apis {
		a := &API{
			Name:       api.Name,
			Namespace:  api.Namespace,
			Labels:     api.Labels,
			PathPrefix: api.Spec.PathPrefix,
			Service: APIService{
				Name: api.Spec.Service.Name,
				Port: APIServiceBackendPort{
					Name:   api.Spec.Service.Port.Name,
					Number: api.Spec.Service.Port.Number,
				},
				OpenAPISpec: OpenAPISpec{
					URL:      api.Spec.Service.OpenAPISpec.URL,
					Path:     api.Spec.Service.OpenAPISpec.Path,
					Protocol: api.Spec.Service.OpenAPISpec.Protocol,
				},
			},
		}

		if api.Spec.Service.OpenAPISpec.Port != nil {
			a.Service.OpenAPISpec.Port = &APIServiceBackendPort{
				Name:   api.Spec.Service.OpenAPISpec.Port.Name,
				Number: api.Spec.Service.OpenAPISpec.Port.Number,
			}
		}

		result[objectKey(a.Name, a.Namespace)] = a
	}

	return result, nil
}

func (f *Fetcher) getAPIAccesses() (map[string]*APIAccess, error) {
	apiAccesses, err := f.hub.Hub().V1alpha1().APIAccesses().Lister().List(labels.Everything())
	if err != nil {
		return nil, err
	}

	result := make(map[string]*APIAccess)
	for _, apiAccess := range apiAccesses {
		a := &APIAccess{
			Name:                  apiAccess.Name,
			Labels:                apiAccess.Labels,
			Groups:                apiAccess.Spec.Groups,
			APISelector:           apiAccess.Spec.APISelector,
			APICollectionSelector: apiAccess.Spec.APICollectionSelector,
		}

		result[a.Name] = a
	}

	return result, nil
}
