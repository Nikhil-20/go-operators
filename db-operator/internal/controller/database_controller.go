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
	"time"

	databasev1 "github.com/nikhil-20/db-operator/api/v1"
	"github.com/nikhil-20/db-operator/internal/secret"
	"github.com/nikhil-20/db-operator/internal/service"
	"github.com/nikhil-20/db-operator/internal/statefulset"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// DatabaseReconciler reconciles a Database object
type DatabaseReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

const (
	finalizerName = "dboperator/finalizer"
)

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

	if !controllerutil.ContainsFinalizer(database, finalizerName) {
		controllerutil.AddFinalizer(database, finalizerName)
		if err := r.Update(ctx, database); err != nil {
			return ctrl.Result{}, nil
		}
		logger.Info("Added finalizer", "name", finalizerName)
	}

	if !database.DeletionTimestamp.IsZero() {
		return r.handleDeletion(ctx, database)
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

func (r *DatabaseReconciler) handleDeletion(ctx context.Context, db *databasev1.Database) (ctrl.Result, error) {

	logger := log.FromContext(ctx)

	if !controllerutil.ContainsFinalizer(db, finalizerName) {
		logger.Info("Finalizer not found")
		return ctrl.Result{}, nil
	}

	logger.Info("handling deletion", "name", finalizerName)

	if err := r.cleanupExternalResources(ctx, db); err != nil {
		logger.Error(err, "Failed to cleanup external resources")

		condition := metav1.Condition{
			Type:               "Ready",
			Status:             metav1.ConditionFalse,
			Reason:             "CleanupFailed",
			Message:            err.Error(),
			LastTransitionTime: metav1.Now(),
			ObservedGeneration: db.Generation,
		}

		meta.SetStatusCondition(&db.Status.Conditions, condition)
		r.Status().Update(ctx, db)
		return ctrl.Result{RequeueAfter: 10 * time.Second}, err
	}

	//cleanup success
	controllerutil.RemoveFinalizer(db, finalizerName)

	if err := r.Update(ctx, db); err != nil {
		return ctrl.Result{}, err
	}

	logger.Info("Finalizer removed, resource will be deleted")
	return ctrl.Result{}, nil
}

func (r *DatabaseReconciler) cleanupExternalResources(ctx context.Context, db *databasev1.Database) error {

	logger := log.FromContext(ctx)
	//deleting statefulset
	stf := &appsv1.StatefulSet{}
	err := r.Get(ctx, client.ObjectKey{
		Name:      db.Name,
		Namespace: db.Namespace,
	}, stf)

	if err == nil {
		// StatefulSet exists, delete it
		logger.Info("Deleting StatefulSet", "name", stf.Name)
		if err := r.Delete(ctx, stf); err != nil && errors.IsNotFound(err) {
			return fmt.Errorf("failed to delete StatefulSet: %w", err)
		}
		// Requeue to wait for deletion to complete
		return fmt.Errorf("waiting for StatefulSet to be deleted")
	} else if !errors.IsNotFound(err) {
		// Some other error occurred
		return fmt.Errorf("failed to get StatefulSet: %w", err)
	}

	//  now cleanup Service

	service := &corev1.Service{}
	err = r.Get(ctx, client.ObjectKey{
		Name:      db.Name,
		Namespace: db.Namespace,
	}, service)

	if err == nil {
		logger.Info("Deleting Service", "name", service.Name)
		if err := r.Delete(ctx, service); err != nil && !errors.IsNotFound(err) {
			return fmt.Errorf("failed to delete Service: %w", err)
		}
		return fmt.Errorf("waiting for Service to be deleted")
	} else if !errors.IsNotFound(err) {
		return fmt.Errorf("failed to get Service: %w", err)
	}

	// Cleaning Secret
	secretname := &corev1.Secret{}
	err = r.Get(ctx, client.ObjectKey{
		Name:      secret.GetName(db),
		Namespace: db.Namespace,
	}, secretname)

	if err == nil {
		logger.Info("Deleting Secret", "name", secretname.Name)
		if err := r.Delete(ctx, secretname); err != nil && !errors.IsNotFound(err) {
			return fmt.Errorf("failed to delete Secret: %w", err)
		}
		return fmt.Errorf("waiting for Secret to be deleted")
	} else if !errors.IsNotFound(err) {
		return fmt.Errorf("failed to get Secret: %w", err)
	}

	// Example: Delete backup in external system
	// if err := r.deleteBackup(ctx, db); err != nil {
	//     return err
	// }

	logger.Info("Cleanup completed")
	return nil

}

func (r *DatabaseReconciler) updateStatus(ctx context.Context, db *databasev1.Database) error {
	// Set the secret name in status
	db.Status.SecretName = secret.GetName(db)

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
