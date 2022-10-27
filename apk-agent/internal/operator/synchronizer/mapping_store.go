package synchronizer

import (
	"sync"

	"k8s.io/apimachinery/pkg/types"
)

type APIToHttpRoutesMapping struct {
	mu sync.Mutex

	mappings map[types.NamespacedName]types.NamespacedName
}

func NewAPIMappingDataStore() *APIToHttpRoutesMapping {
	return &APIToHttpRoutesMapping{
		mappings: map[types.NamespacedName]types.NamespacedName{},
	}
}

func (apiHttp *APIToHttpRoutesMapping) GetHttpRouteForAPI(api types.NamespacedName) types.NamespacedName {
	apiHttp.mu.Lock()
	defer apiHttp.mu.Unlock()

	return apiHttp.mappings[api]
}

func (apiHttp *APIToHttpRoutesMapping) UpdateHttpRouteForAPI(api types.NamespacedName, httpRoute types.NamespacedName) error {
	apiHttp.mu.Lock()
	defer apiHttp.mu.Unlock()

	// _, ok := apiHttp.mappings[api]
	// if !ok {
	// 	apiHttp.mappings[api] = httpRoute
	// }
	apiHttp.mappings[api] = httpRoute
	return nil
}

func (apiHttp *APIToHttpRoutesMapping) DeleteHttpRouteForAPI(api types.NamespacedName, httpRoute types.NamespacedName) {
	apiHttp.mu.Lock()
	defer apiHttp.mu.Unlock()

	delete(apiHttp.mappings, api)
}

type HttpRoutesToAPIMapping struct {
	mu sync.Mutex

	mappings map[types.NamespacedName]types.NamespacedName
}

func NewHttpRouteMappingDataStore() *HttpRoutesToAPIMapping {
	return &HttpRoutesToAPIMapping{
		mappings: map[types.NamespacedName]types.NamespacedName{},
	}
}

func (httpAPI *HttpRoutesToAPIMapping) GetAPIForHttpRoute(httpRoute types.NamespacedName) types.NamespacedName {
	httpAPI.mu.Lock()
	defer httpAPI.mu.Unlock()

	return httpAPI.mappings[httpRoute]
}

func (httpAPI *HttpRoutesToAPIMapping) UpdateAPIForHttpRoute(httpRoute types.NamespacedName, api types.NamespacedName) error {
	httpAPI.mu.Lock()
	defer httpAPI.mu.Unlock()

	httpAPI.mappings[httpRoute] = api
	return nil
}

func (httpAPI *HttpRoutesToAPIMapping) DeleteAPIForHttpRoute(httpRoute types.NamespacedName) error {
	httpAPI.mu.Lock()
	defer httpAPI.mu.Unlock()

	delete(httpAPI.mappings, httpRoute)
	return nil
}
