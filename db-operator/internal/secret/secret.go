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

// Package secret handles reconciliation of the credentials Secret for a Database.
package secret

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"

	databasev1 "github.com/nikhil-20/db-operator/api/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// Name returns the name of the Secret for this Database
func Name(db *databasev1.Database) string {
	return fmt.Sprintf("%s-credentials", db.Name)
}

// generatePassword generates a random password
func generatePassword(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes)[:length], nil
}

// Reconcile ensures the credentials Secret exists
func ReconcileSecret(ctx context.Context, c client.Client, scheme *runtime.Scheme, db *databasev1.Database) error {
	logger := log.FromContext(ctx)
	secretName := Name(db)

	sec := &corev1.Secret{}
	err := c.Get(ctx, client.ObjectKey{
		Name:      secretName,
		Namespace: db.Namespace,
	}, sec)

	if errors.IsNotFound(err) {
		// Generate random password
		password, err := generatePassword(16)
		if err != nil {
			return fmt.Errorf("failed to generate password: %w", err)
		}

		// Create new secret
		sec = &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      secretName,
				Namespace: db.Namespace,
			},
			Type: corev1.SecretTypeOpaque,
			StringData: map[string]string{
				"username": db.Spec.Username,
				"password": password,
				"database": db.Spec.DatabaseName,
			},
		}

		// Set owner reference
		if err := ctrl.SetControllerReference(db, sec, scheme); err != nil {
			return err
		}

		logger.Info("Creating Secret", "name", secretName)
		return c.Create(ctx, sec)
	} else if err != nil {
		return err
	}

	// Secret already exists, don't update password
	return nil
}
