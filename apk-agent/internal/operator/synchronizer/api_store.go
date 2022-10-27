package synchronizer

import (
	"sync"

	"github.com/AmaliMatharaarachchi/APKAgent/apk-agent/internal/logger"
	"github.com/AmaliMatharaarachchi/APKAgent/apk-agent/internal/operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/types"
	gwapiv1b1 "sigs.k8s.io/gateway-api/apis/v1beta1"
)

type APIStore struct {
	mu sync.Mutex

	mappings map[types.NamespacedName]*API
}

func NewAPIStore() *APIStore {
	return &APIStore{
		mappings: map[types.NamespacedName]*API{},
	}
}

func (apiS *APIStore) StoreNewAPI(apiName types.NamespacedName, api API) error {
	apiS.mu.Lock()
	defer apiS.mu.Unlock()

	apiS.mappings[apiName] = &api
	logger.LoggerOperator.Infof("API: %v added to API store", apiName)
	return nil
}

func (apiS *APIStore) AddHttpRouteToAPI(apiName types.NamespacedName, httpRoute *gwapiv1b1.HTTPRoute, production bool) error {
	apiS.mu.Lock()
	defer apiS.mu.Unlock()

	if production {
		apiS.mappings[apiName].ProdHttpRoute = httpRoute
	} else {
		apiS.mappings[apiName].SandHttpRoute = httpRoute
	}
	return nil
}

func (apiS *APIStore) LoadStoredAPI(apiName types.NamespacedName) (API, bool) {
	apiS.mu.Lock()
	defer apiS.mu.Unlock()

	api, ok := apiS.mappings[apiName]
	if !ok {
		return API{}, ok
	}
	return *api, ok
}

func (apiS *APIStore) UpdateAPIDefinition(apiName types.NamespacedName, apiDef *v1alpha1.API) error {
	apiS.mu.Lock()
	defer apiS.mu.Unlock()

	_, ok := apiS.mappings[apiName]
	if !ok {
		apiS.mappings[apiName] = &API{
			APIDefinition: apiDef,
		}
		return nil
	}
	apiS.mappings[apiName].APIDefinition = apiDef
	return nil
}
