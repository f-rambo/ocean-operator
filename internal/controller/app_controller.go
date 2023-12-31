/*
Copyright 2023 f-rambo.

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

package controller

import (
	"context"

	log "github.com/f-rambo/operatorapp/utils/log"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	operatoroceaniov1alpha1 "github.com/f-rambo/operatorapp/api/v1alpha1"
)

// AppReconciler reconciles a App object
type AppReconciler struct {
	client.Client
	Cfg      *rest.Config
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
	Log      *log.Helper
}

//+kubebuilder:rbac:groups=operator.ocean.io,resources=apps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=operator.ocean.io,resources=apps/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=operator.ocean.io,resources=apps/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the App object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *AppReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	app := &operatoroceaniov1alpha1.App{}
	err := r.Get(ctx, req.NamespacedName, app)
	if err != nil {
		if errors.IsNotFound(err) {
			err = r.deleteApp(ctx, req.Name)
			if err != nil {
				r.Recorder.Eventf(app, corev1.EventTypeWarning, "delete app", "Failed to delete App: %v", err)
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	if app.Spec.AppChart.Enable {
		err = r.deployAppChart(ctx, req, app)
		if err != nil {
			r.Log.Error(err, "deployAppChart")
			r.Recorder.Eventf(app, corev1.EventTypeWarning, "deployAppChart", "%v", err)
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}
	if app.Spec.Service.Enable {
		err = r.deployServiceApp(ctx, req, app)
		if err != nil {
			r.Log.Error(err, "deployServiceApp")
			r.Recorder.Eventf(app, corev1.EventTypeWarning, "deployServiceApp", "%v", err)
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *AppReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&operatoroceaniov1alpha1.App{}).
		Complete(r)
}
