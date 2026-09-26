package statefulset

import (
	"context"
	"fmt"
	"time"

	databasev1 "github.com/nikhil-20/db-operator/api/v1"
	"github.com/nikhil-20/db-operator/internal/secret"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func ReconcileSFS(ctx context.Context, cli client.Client, scheme *runtime.Scheme, db *databasev1.Database) error {

	logger := log.FromContext(ctx)

	statefulSet := &appsv1.StatefulSet{}
	err := cli.Get(ctx, client.ObjectKey{
		Name:      db.Name,
		Namespace: db.Namespace,
	}, statefulSet)

	desiredStatefulSet := buildStatefulSet(db)

	if errors.IsNotFound(err) {
		// Set owner reference
		if err := ctrl.SetControllerReference(db, desiredStatefulSet, scheme); err != nil {
			return err
		}
		logger.Info("Creating StatefulSet", "name", desiredStatefulSet.Name)
		return cli.Create(ctx, desiredStatefulSet)
	} else if err != nil {
		return err
	}

	if statefulSet.Spec.Replicas != desiredStatefulSet.Spec.Replicas {
		return patchStatefulSetReplicas(ctx, statefulSet, cli, *desiredStatefulSet.Spec.Replicas)
	}

	// Update if needed
	if statefulSet.Spec.Replicas != desiredStatefulSet.Spec.Replicas ||
		statefulSet.Spec.Template.Spec.Containers[0].Image != desiredStatefulSet.Spec.Template.Spec.Containers[0].Image {
		statefulSet.Spec = desiredStatefulSet.Spec
		logger.Info("Updating StatefulSet", "name", statefulSet.Name)
		// Use retry logic for updates
		return updateWithRetry(ctx, cli, statefulSet, 3)
	}

	return nil
}

func patchStatefulSetReplicas(ctx context.Context, statefulSet *appsv1.StatefulSet, cli client.Client, replicas int32) error {
	patch := client.MergeFrom(statefulSet.DeepCopy())
	statefulSet.Spec.Replicas = &replicas
	return cli.Patch(ctx, statefulSet, patch)
}

func buildStatefulSet(db *databasev1.Database) *appsv1.StatefulSet {
	replicas := int32(1)
	if db.Spec.Replicas != nil {
		replicas = *db.Spec.Replicas
	}

	image := db.Spec.Image
	if image == "" {
		image = "postgres:14"
	}

	secretName := secret.Name(db)

	return &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      db.Name,
			Namespace: db.Namespace,
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app":      "database",
					"database": db.Name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app":      "database",
						"database": db.Name,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "postgres",
							Image: image,
							Env: []corev1.EnvVar{
								{
									Name:  "POSTGRES_DB",
									Value: db.Spec.DatabaseName,
								},
								{
									Name: "POSTGRES_USER",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: secretName,
											},
											Key: "username",
										},
									},
								},
								{
									Name: "POSTGRES_PASSWORD",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: secretName,
											},
											Key: "password",
										},
									},
								},
								{
									Name:  "PGDATA",
									Value: "/var/lib/postgresql/data/pgdata",
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "data",
									MountPath: "/var/lib/postgresql/data",
								},
							},
						},
					},
				},
			},
			VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "data",
					},
					Spec: corev1.PersistentVolumeClaimSpec{
						AccessModes: []corev1.PersistentVolumeAccessMode{
							corev1.ReadWriteOnce,
						},
						Resources: corev1.VolumeResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceStorage: resource.MustParse(db.Spec.Storage.Size),
							},
						},
					},
				},
			},
		},
	}
}

func updateWithRetry(ctx context.Context, cli client.Client, obj client.Object, maxRetries int) error {
	for i := 0; i < maxRetries; i++ {
		err := cli.Update(ctx, obj)
		if err == nil {
			return nil
		}

		if !errors.IsConflict(err) {
			return err
		}

		// Conflict - get fresh version and retry
		key := client.ObjectKeyFromObject(obj)
		if err := cli.Get(ctx, key, obj); err != nil {
			return err
		}

		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("max retries exceeded")
}
