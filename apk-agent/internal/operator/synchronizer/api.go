package synchronizer

import (
	"github.com/AmaliMatharaarachchi/APKAgent/apk-agent/internal/operator/api/v1alpha1"
	gwapiv1b1 "sigs.k8s.io/gateway-api/apis/v1beta1"
)

type API struct {
	APIDefinition *v1alpha1.API
	HttpRoutes    []*gwapiv1b1.HTTPRoute
	APIPolicies   []string
}
