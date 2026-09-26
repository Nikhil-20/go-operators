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

// Package service handles reconciliation of the Service for a Database.
package service

import (
	"context"

	databasev1 "github.com/nikhil-20/db-operator/api/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Build returns the desired Service for this Database
func Build(db *databasev1.Database) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      db.Name,
			Namespace: db.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app":      "database",
				"database": db.Name,
			},
			Ports: []corev1.ServicePort{
				{
					Port: 5432,
					Name: "postgres",
				},
			},
		},
	}
}

// Reconcile ensures the Service exists for this Database
func ReconcileService(ctx context.Context, c client.Client, scheme *runtime.Scheme, db *databasev1.Database) error {
	svc := &corev1.Service{}
	err := c.Get(ctx, client.ObjectKey{
		Name:      db.Name,
		Namespace: db.Namespace,
	}, svc)

	desiredService := Build(db)

	if errors.IsNotFound(err) {
		if err := ctrl.SetControllerReference(db, desiredService, scheme); err != nil {
			return err
		}
		return c.Create(ctx, desiredService)
	} else if err != nil {
		return err
	}

	// Service updates are less common, but handle if needed
	return nil
}
