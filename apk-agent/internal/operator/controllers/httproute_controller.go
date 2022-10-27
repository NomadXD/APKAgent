package controllers

import (
	"context"

	"github.com/AmaliMatharaarachchi/APKAgent/apk-agent/internal/logger"
	"github.com/AmaliMatharaarachchi/APKAgent/apk-agent/internal/operator/synchronizer"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/source"
	gwapiv1b1 "sigs.k8s.io/gateway-api/apis/v1beta1"
)

type HttpRouteReconciler struct {
	client.Client
	//Scheme *runtime.Scheme
	apiStore               *synchronizer.APIStore
	apiToHttpRoutesMapping *synchronizer.APIToHttpRoutesMapping
}

func NewHttpRouteController(mgr manager.Manager, apiStore *synchronizer.APIStore, apiMappingStore *synchronizer.APIToHttpRoutesMapping) error {
	r := &HttpRouteReconciler{
		Client:                 mgr.GetClient(),
		apiStore:               apiStore,
		apiToHttpRoutesMapping: apiMappingStore,
	}
	c, err := controller.New("HttpRoute", mgr, controller.Options{Reconciler: r})
	if err != nil {
		logger.LoggerOperator.Errorf("Error creating HttpRoute Controller: %v\n", err)
		return err
	}

	if err := c.Watch(&source.Kind{Type: &gwapiv1b1.HTTPRoute{}}, &handler.EnqueueRequestForObject{}); err != nil {
		return err
	}
	logger.LoggerOperator.Info("HttpRoute Controller successfully started. Watching HttpRoute Objects....")
	return nil

}

func (r *HttpRouteReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var httpRoute gwapiv1b1.HTTPRoute
	if err := r.Get(ctx, req.NamespacedName, &httpRoute); err != nil {
		logger.LoggerOperator.Errorf("Error fetching HttpRoute: %v", err)
		return ctrl.Result{}, err
	}
	logger.LoggerOperator.Infof("Reconciled HttpRoute: %v", httpRoute.Name)
	return ctrl.Result{}, nil
}
