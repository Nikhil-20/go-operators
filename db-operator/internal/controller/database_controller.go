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

	databasev1 "github.com/nikhil-20/db-operator/api/v1"
	"github.com/nikhil-20/db-operator/internal/secret"
	"github.com/nikhil-20/db-operator/internal/service"
	"github.com/nikhil-20/db-operator/internal/statefulset"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// DatabaseReconciler reconciles a Database object
type DatabaseReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=database.db-operator.io,resources=databases,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=database.db-operator.io,resources=databases/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=database.db-operator.io,resources=databases/finalizers,verbs=update
// +kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;patch;delete
func (r *DatabaseReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := logf.FromContext(ctx)

	database := &databasev1.Database{}

	if err := r.Get(ctx, req.NamespacedName, database); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, err
	}

	logger.Info("Reconciling Database", "name", database.Name)

	// Reconcile Secret (must be done before StatefulSet)
	if err := secret.ReconcileSecret(ctx, r.Client, r.Scheme, database); err != nil {
		return ctrl.Result{}, err
	}

	// Reconcile StatefulSet
	if err := statefulset.ReconcileSFS(ctx, r.Client, r.Scheme, database); err != nil {
		return ctrl.Result{}, err
	}

	// Reconcile Service
	if err := service.ReconcileService(ctx, r.Client, r.Scheme, database); err != nil {
		return ctrl.Result{}, err
	}

	// Update status
	if err := r.updateStatus(ctx, database); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *DatabaseReconciler) updateStatus(ctx context.Context, db *databasev1.Database) error {
	// Set the secret name in status
	db.Status.SecretName = secret.Name(db)

	// Check StatefulSet status
	statefulSet := &appsv1.StatefulSet{}
	err := r.Get(ctx, client.ObjectKey{
		Name:      db.Name,
		Namespace: db.Namespace,
	}, statefulSet)

	if err != nil {
		db.Status.Phase = "Pending"
		db.Status.Ready = false
	} else {
		if statefulSet.Status.ReadyReplicas == *statefulSet.Spec.Replicas {
			db.Status.Phase = "Ready"
			db.Status.Ready = true
			db.Status.Endpoint = fmt.Sprintf("%s.%s.svc.cluster.local:5432", db.Name, db.Namespace)
		} else {
			db.Status.Phase = "Creating"
			db.Status.Ready = false
		}
	}

	return r.Status().Update(ctx, db)
}

// SetupWithManager sets up the controller with the Manager.
func (r *DatabaseReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&databasev1.Database{}).
		Named("databaseoperator").
		Owns(&appsv1.StatefulSet{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.Secret{}).
		Complete(r)
}
