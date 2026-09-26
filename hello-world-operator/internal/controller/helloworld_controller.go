/*
Copyright 2026.

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
	"fmt"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	hellov1 "github.com/example/hello-world-operator/api/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// HelloWorldReconciler reconciles a HelloWorld object
type HelloWorldReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=hello.example.com,resources=helloworlds,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=hello.example.com,resources=helloworlds/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=hello.example.com,resources=helloworlds/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the HelloWorld object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.25.0/pkg/reconcile
func (r *HelloWorldReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := logf.FromContext(ctx)

	logger.Info("Reconcilation in Helloworld", "name", req.NamespacedName)

	hello := &hellov1.HelloWorld{}

	if err := r.Get(ctx, req.NamespacedName, hello); err != nil {
		if errors.IsNotFound(err) {
			logger.Info("Hello world instance not found, ignoring", "name",
				req.NamespacedName)

			return ctrl.Result{}, nil
		}
		logger.Error(err, "Hello world not found")
		return ctrl.Result{}, err
	}

	configMapName := hello.Name + "-config"
	configMap := &corev1.ConfigMap{

		ObjectMeta: metav1.ObjectMeta{
			Name:      configMapName,
			Namespace: hello.Namespace,
		},
		Data: map[string]string{
			"message": hello.Spec.Message,
			"count":   fmt.Sprintf("%d", hello.Spec.Count),
		},
	}

	//set owner reference
	if err := ctrl.SetControllerReference(hello, configMap, r.Scheme); err != nil {
		logger.Error(err, "Failed to set controller reference")
		return ctrl.Result{}, err
	}

	existingConfigmap := &corev1.ConfigMap{}
	err := r.Get(ctx, client.ObjectKey{
		Name:      configMap.Name,
		Namespace: configMap.Namespace,
	}, existingConfigmap)

	if err != nil && errors.IsNotFound(err) {

		// ConfigMap doesn't exist, create it
		logger.Info("Creating ConfigMap", "name", configMap.Name)
		if err := r.Create(ctx, configMap); err != nil {
			logger.Error(err, "Failed to create ConfigMap")
			return ctrl.Result{}, err
		} else if err != nil {
			logger.Error(err, "Failed to get ConfigMap")
			return ctrl.Result{}, err
		} else {
			if existingConfigmap.Data["message"] != configMap.Data["message"] ||
				existingConfigmap.Data["count"] != configMap.Data["count"] {
				logger.Info("Updating ConfigMap", "name", configMap.Name)
				existingConfigmap.Data = configMap.Data
				if err := r.Update(ctx, existingConfigmap); err != nil {
					logger.Error(err, "Failed to update ConfigMap")
					return ctrl.Result{}, err
				}
			}
		}
	}

	//update status
	now := metav1.Now()
	hello.Status.Phase = "Ready"
	hello.Status.ConfigMapCreated = true
	hello.Status.LastUpdated = &now

	if err := r.Status().Update(ctx, hello); err != nil {
		logger.Error(err, "Failed to update status")
		return ctrl.Result{}, err
	}

	logger.Info("Successfully reconciled HelloWorld", "name", req.NamespacedName)
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *HelloWorldReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&hellov1.HelloWorld{}).
		Named("helloworld").
		Complete(r)
}
