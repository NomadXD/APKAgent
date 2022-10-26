/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/AmaliMatharaarachchi/APKAgent/apk-agent/internal/logger"
	dpv1alpha1 "github.com/AmaliMatharaarachchi/APKAgent/apk-agent/internal/operator/api/v1alpha1"
	gwapiv1b1 "sigs.k8s.io/gateway-api/apis/v1beta1"
)

// APIReconciler reconciles a API object
type APIReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=dp.wso2.com,resources=apis,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=dp.wso2.com,resources=apis/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=dp.wso2.com,resources=apis/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the API object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *APIReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)
	logger.LoggerOperator.Infof("Log from context: %v\n", l)
	logger.LoggerOperator.Infof("Log from request: %v\n", req)
	var api dpv1alpha1.API
	if err := r.Get(ctx, req.NamespacedName, &api); err != nil {
		logger.LoggerOperator.Errorf("Error fetching the API from Kubernetes API server: %v\n", err)
	}

	// logger.LoggerOperator.Infof("Log from reconciler: %v\n", api)

	var httpRoutes []*gwapiv1b1.HTTPRoute
	for _, httpRouteRef := range api.Spec.ProdHTTPRouteRefs {
		var httpRoute gwapiv1b1.HTTPRoute
		logger.LoggerOperator.Infof("HttpRouteRef: %v\n", httpRouteRef)
		if err := r.Get(ctx, types.NamespacedName{Name: httpRouteRef, Namespace: req.Namespace}, &httpRoute); err != nil {
			logger.LoggerOperator.Errorf("Error fetching the HttpRoute: %v for API: %v\n", httpRouteRef, api.Name)
			return ctrl.Result{}, err
		}
		httpRoutes = append(httpRoutes, &httpRoute)
	}
	logger.LoggerOperator.Infof("Http Routes: %v\n", httpRoutes)
	
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *APIReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&dpv1alpha1.API{}).
		Complete(r)
}
