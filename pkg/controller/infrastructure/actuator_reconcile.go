// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package infrastructure

import (
	"bytes"
	"context"
	"fmt"

	extensionscontroller "github.com/gardener/gardener/extensions/pkg/controller"
	"github.com/gardener/gardener/extensions/pkg/terraformer"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	apiv1alpha1 "github.com/gardener/gardener-extension-provider-equinix-metal/pkg/apis/equinixmetal/v1alpha1"
	"github.com/gardener/gardener-extension-provider-equinix-metal/pkg/equinixmetal"
)

func (a *actuator) Reconcile(ctx context.Context, log logr.Logger, infrastructure *extensionsv1alpha1.Infrastructure, cluster *extensionscontroller.Cluster) error {
	log.WithValues("infrastructure", client.ObjectKeyFromObject(infrastructure), "operation", "reconcile")
	return a.reconcile(ctx, log, infrastructure, cluster, terraformer.StateConfigMapInitializerFunc(terraformer.CreateState))
}

func (a *actuator) reconcile(ctx context.Context, logger logr.Logger, infrastructure *extensionsv1alpha1.Infrastructure, cluster *extensionscontroller.Cluster, stateInitializer terraformer.StateConfigMapInitializer) error {
	var (
		terraformConfig = GenerateTerraformInfraConfig(infrastructure)
		mainTF          bytes.Buffer
	)

	if err := tplMainTF.Execute(&mainTF, terraformConfig); err != nil {
		return fmt.Errorf("could not render Terraform template: %+v", err)
	}

	tf, err := a.newTerraformer(logger, equinixmetal.TerraformerPurposeInfra, infrastructure)
	if err != nil {
		return fmt.Errorf("could not create terraformer object: %+v", err)
	}

	if err := tf.
		SetEnvVars(generateTerraformInfraVariablesEnvironment(infrastructure.Spec.SecretRef)...).
		InitializeWith(ctx,
			terraformer.DefaultInitializer(
				a.client,
				mainTF.String(),
				variablesTF,
				[]byte(terraformTFVars),
				stateInitializer,
			)).
		Apply(ctx); err != nil {

		return errors.Wrap(err, "failed to apply the terraform config")
	}

	return a.updateProviderStatus(ctx, tf, infrastructure)
}

// GenerateTerraformInfraConfig generates the Equinix Metal Terraform configuration based on the given infrastructure and project.
func GenerateTerraformInfraConfig(infrastructure *extensionsv1alpha1.Infrastructure) map[string]interface{} {
	return map[string]interface{}{
		"sshPublicKey": string(infrastructure.Spec.SSHPublicKey),
		"clusterName":  infrastructure.Namespace,
		"outputKeys": map[string]interface{}{
			"sshKeyID": equinixmetal.SSHKeyID,
		},
	}
}

func (a *actuator) updateProviderStatus(ctx context.Context, tf terraformer.Terraformer, infrastructure *extensionsv1alpha1.Infrastructure) error {
	outputVarKeys := []string{
		equinixmetal.SSHKeyID,
	}

	output, err := tf.GetStateOutputVariables(ctx, outputVarKeys...)
	if err != nil {
		return err
	}

	state, err := tf.GetRawState(ctx)
	if err != nil {
		return err
	}
	stateByte, err := state.Marshal()
	if err != nil {
		return err
	}

	patch := client.MergeFrom(infrastructure.DeepCopy())
	infrastructure.Status.ProviderStatus = &runtime.RawExtension{
		Object: &apiv1alpha1.InfrastructureStatus{
			TypeMeta: metav1.TypeMeta{
				APIVersion: apiv1alpha1.SchemeGroupVersion.String(),
				Kind:       "InfrastructureStatus",
			},
			SSHKeyID: output[equinixmetal.SSHKeyID],
		},
	}
	infrastructure.Status.State = &runtime.RawExtension{Raw: stateByte}
	return a.client.Status().Patch(ctx, infrastructure, patch)
}
