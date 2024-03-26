// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package infrastructure

import (
	"time"

	"github.com/gardener/gardener/extensions/pkg/controller/infrastructure"
	"github.com/gardener/gardener/extensions/pkg/terraformer"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/gardener/gardener-extension-provider-equinix-metal/imagevector"
	"github.com/gardener/gardener-extension-provider-equinix-metal/pkg/equinixmetal"
)

type actuator struct {
	client                     client.Client
	restConfig                 *rest.Config
	disableProjectedTokenMount bool
}

// NewActuator creates a new Actuator that updates the status of the handled Infrastructure resources.
func NewActuator(
	mgr manager.Manager,
	disableProjectedTokenMount bool,
) infrastructure.Actuator {
	return &actuator{
		restConfig:                 mgr.GetConfig(),
		client:                     mgr.GetClient(),
		disableProjectedTokenMount: disableProjectedTokenMount,
	}
}

// Helper functions

func (a *actuator) newTerraformer(logger logr.Logger, purpose string, infra *extensionsv1alpha1.Infrastructure) (terraformer.Terraformer, error) {
	tf, err := terraformer.NewForConfig(logger, a.restConfig, purpose, infra.GetNamespace(), infra.GetName(), imagevector.TerraformerImage())
	if err != nil {
		return nil, err
	}

	owner := metav1.NewControllerRef(infra, extensionsv1alpha1.SchemeGroupVersion.WithKind(extensionsv1alpha1.InfrastructureResource))
	return tf.
		UseProjectedTokenMount(!a.disableProjectedTokenMount).
		SetTerminationGracePeriodSeconds(630).
		SetDeadlineCleaning(5 * time.Minute).
		SetDeadlinePod(15 * time.Minute).
		SetOwnerRef(owner), nil
}

func generateTerraformInfraVariablesEnvironment(secretRef corev1.SecretReference) []corev1.EnvVar {
	return []corev1.EnvVar{
		{
			Name: "TF_VAR_EQXM_API_KEY",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: secretRef.Name,
					},
					Key: equinixmetal.APIToken,
				},
			},
		},
		{
			Name: "TF_VAR_EQXM_PROJECT_ID",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: secretRef.Name,
					},
					Key: equinixmetal.ProjectID,
				},
			},
		},
	}
}
