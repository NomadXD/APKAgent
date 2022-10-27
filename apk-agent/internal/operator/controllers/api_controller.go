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

	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/AmaliMatharaarachchi/APKAgent/apk-agent/internal/logger"
	dpv1alpha1 "github.com/AmaliMatharaarachchi/APKAgent/apk-agent/internal/operator/api/v1alpha1"
	"github.com/AmaliMatharaarachchi/APKAgent/apk-agent/internal/operator/synchronizer"
	"github.com/AmaliMatharaarachchi/APKAgent/apk-agent/internal/operator/utils"
	gwapiv1b1 "sigs.k8s.io/gateway-api/apis/v1beta1"
)

// APIReconciler reconciles a API object
type APIReconciler struct {
	client.Client
	//Scheme *runtime.Scheme
	apiStore               *synchronizer.APIStore
	apiToHttpRoutesMapping *synchronizer.APIToHttpRoutesMapping
	httpRouteToAPIMapping  *synchronizer.HttpRoutesToAPIMapping
	ch                     *chan string
}

func NewAPIController(mgr manager.Manager, apiStore *synchronizer.APIStore, apiMappingStore *synchronizer.APIToHttpRoutesMapping,
	httpRouteMappingStore *synchronizer.HttpRoutesToAPIMapping, ch *chan string) error {
	r := &APIReconciler{
		Client:                 mgr.GetClient(),
		apiStore:               apiStore,
		apiToHttpRoutesMapping: apiMappingStore,
		httpRouteToAPIMapping:  httpRouteMappingStore,

		ch: ch,
	}
	c, err := controller.New("API", mgr, controller.Options{Reconciler: r})
	if err != nil {
		logger.LoggerOperator.Errorf("Error creating API Controller: %v\n", err)
		return err
	}

	if err := c.Watch(&source.Kind{Type: &dpv1alpha1.API{}}, &handler.EnqueueRequestForObject{}); err != nil {
		return err
	}
	if err := c.Watch(&source.Kind{Type: &gwapiv1b1.HTTPRoute{}}, handler.EnqueueRequestsFromMapFunc(r.getHTTPRoutesForAPI)); err != nil {
		logger.LoggerOperator.Errorf("Error watching HttpRoute from API Controller: %v", err)
		return err
	}
	logger.LoggerOperator.Info("API Controller successfully started. Watching API Objects....")
	return nil
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

	// apiList := &dpv1alpha1.APIList{}
	// if err := r.Client.List(ctx, apiList); err != nil {
	// 	logger.LoggerOperator.Errorf("Error listing APIs from the controller cache: %v", err)
	// 	return reconcile.Result{}, fmt.Errorf("Error listing APIs from the controller cache")
	// }

	// apiFound := false
	// for _, api := range apiList.Items {
	// 	apiKey := utils.NamespacedName(&api)
	// 	if apiKey == req.NamespacedName {
	// 		apiFound = true
	// 	}
	// }
	var api dpv1alpha1.API
	if err := r.Get(ctx, req.NamespacedName, &api); err != nil {
		logger.LoggerOperator.Errorf("Error fetching the API from Kubernetes API server: %v\n", err)
	}

	// logger.LoggerOperator.Infof("Log from reconciler: %v\n", api)

	if err := validateHttpRouteRef(ctx, r.Client, req.Namespace, api.Spec.ProdHTTPRouteRef, api.Spec.SandHTTPRouteRef); err != nil {
		logger.LoggerOperator.Errorf("Error validating httpRouteRefs for the API: %v:%v", api.Spec.APIDisplayName, err)
		return ctrl.Result{}, err
	}

	if apiStored, ok := r.apiStore.LoadStoredAPI(utils.NamespacedName(&api)); !ok || (api.Generation > apiStored.APIDefinition.Generation) {
		r.apiStore.UpdateAPIDefinition(utils.NamespacedName(&api), &api)
		r.apiToHttpRoutesMapping.UpdateHttpRouteForAPI(utils.NamespacedName(&api), types.NamespacedName{Namespace: req.Namespace, Name: api.Spec.ProdHTTPRouteRef})
		r.httpRouteToAPIMapping.UpdateAPIForHttpRoute(types.NamespacedName{Namespace: req.Namespace, Name: api.Spec.ProdHTTPRouteRef}, utils.NamespacedName(&api))
	}

	// if err := r.apiStore.StoreNewAPI(types.NamespacedName{Name: api.Name, Namespace: api.Namespace},
	// 	synchronizer.API{APIDefinition: &api}); err != nil {
	// 	logger.LoggerOperator.Errorf("Error saving the reconciled API in the API store: %v\n", err)
	// }

	// r.getHTTPRoutesForAPI()
	// var prodHttpRoute gwapiv1b1.HTTPRoute
	// if err := r.Get(ctx, types.NamespacedName{Name: api.Spec.ProdHTTPRouteRef, Namespace: req.Namespace}, &prodHttpRoute); err != nil {
	// 	logger.LoggerOperator.Errorf("Error fetching the ProdHttpRoute: %v for API: %v\n", api.Spec.ProdHTTPRouteRef, api.Name)
	// 	return ctrl.Result{}, err
	// }
	// if err := r.apiStore.AddHttpRouteToAPI(types.NamespacedName{Name: req.Name, Namespace: req.Namespace}, &prodHttpRoute, true); err != nil {
	// 	logger.LoggerOperator.Errorf("Error stroing ProdHttpRoute in the API store for API:ProdHttpRoute:%v:%v", api.Name, api.Spec.ProdHTTPRouteRef)
	// }

	// if api.Spec.SandHTTPRouteRef != "" {
	// 	var sandHttpRoute *gwapiv1b1.HTTPRoute
	// 	if err := r.Get(ctx, types.NamespacedName{Name: api.Spec.SandHTTPRouteRef, Namespace: req.Namespace}, sandHttpRoute); err != nil {
	// 		logger.LoggerOperator.Errorf("Error fetching the SandHttpRoute: %v for API: %v\n", api.Spec.SandHTTPRouteRef, api.Name)
	// 		return ctrl.Result{}, err
	// 	}
	// 	if err := r.apiStore.AddHttpRouteToAPI(types.NamespacedName{Name: req.Name, Namespace: req.Namespace}, sandHttpRoute, false); err != nil {
	// 		logger.LoggerOperator.Errorf("Error stroing SandHttpRoute in the API store for API:SandHttpRoute:%v:%v", api.Name, api.Spec.SandHTTPRouteRef)
	// 	}
	// }
	logger.LoggerOperator.Infof("API store: %v\n", r.apiStore)
	*r.ch <- "foo"

	return ctrl.Result{}, nil
}

func validateHttpRouteRef(ctx context.Context, client client.Client, namespace string, prodHttpRouteRef string, sandHttpRouteRef string) error {
	var prodHttpRoute gwapiv1b1.HTTPRoute
	if err := client.Get(ctx, types.NamespacedName{Namespace: namespace, Name: prodHttpRouteRef}, &prodHttpRoute); err != nil {
		logger.LoggerOperator.Errorf("Error fetching ProdHTTPRoute: %v:%v", prodHttpRouteRef, err)
		return err
	}
	if sandHttpRouteRef != "" {
		var sandHttpRoute gwapiv1b1.HTTPRoute
		if err := client.Get(ctx, types.NamespacedName{Namespace: namespace, Name: sandHttpRouteRef}, &sandHttpRoute); err != nil {
			logger.LoggerOperator.Errorf("Error fetching SandHTTPRoute: %v:%v", sandHttpRouteRef, err)
			return err
		}
	}
	return nil
}

func (r *APIReconciler) getHTTPRoutesForAPI(obj client.Object) []reconcile.Request {
	httpRoute, ok := obj.(*gwapiv1b1.HTTPRoute)
	if !ok {
		logger.LoggerOperator.Errorf("Unexpected object type, bypassing reconciliation: %v", obj)
	}
	logger.LoggerOperator.Infof("HttpRoute: %v", httpRoute.Name)
	api := r.httpRouteToAPIMapping.GetAPIForHttpRoute(types.NamespacedName{Name: httpRoute.Name, Namespace: httpRoute.Namespace})
	logger.LoggerOperator.Infof("API for HttpRoute: %v: %v", httpRoute.Name, api)
	apiString := api.String()
	logger.LoggerOperator.Info(apiString)
	if api.String() == "/" {
		return []reconcile.Request{}
	}
	requests := []reconcile.Request{}
	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      api.Name,
			Namespace: api.Namespace,
		},
	}
	requests = append(requests, req)
	logger.LoggerOperator.Infof("Reconciliation request created: %v", req)
	return requests
}

// SetupWithManager sets up the controller with the Manager.
// func (r *APIReconciler) SetupWithManager(mgr ctrl.Manager) error {
// 	return ctrl.NewControllerManagedBy(mgr).
// 		For(&dpv1alpha1.API{}).
// 		Complete(r)
// }
