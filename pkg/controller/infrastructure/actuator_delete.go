// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package infrastructure

import (
	"context"
	"fmt"

	extensionscontroller "github.com/gardener/gardener/extensions/pkg/controller"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/gardener/gardener-extension-provider-equinix-metal/pkg/equinixmetal"
)

func (a *actuator) Delete(ctx context.Context, log logr.Logger, infrastructure *extensionsv1alpha1.Infrastructure, cluster *extensionscontroller.Cluster) error {
	log.WithValues("infrastructure", client.ObjectKeyFromObject(infrastructure), "operation", "delete")
	tf, err := a.newTerraformer(log, equinixmetal.TerraformerPurposeInfra, infrastructure)
	if err != nil {
		return fmt.Errorf("could not create the Terraformer: %+v", err)
	}

	// terraform pod from previous reconciliation might still be running, ensure they are gone before doing any operations
	if err := tf.EnsureCleanedUp(ctx); err != nil {
		return err
	}

	// If the Terraform state is empty then we can exit early as we didn't create anything. Though, we clean up potentially
	// created configmaps/secrets related to the Terraformer.
	stateIsEmpty := tf.IsStateEmpty(ctx)
	if stateIsEmpty {
		log.Info("exiting early as infrastructure state is empty - nothing to do")
		return tf.CleanupConfiguration(ctx)
	}

	return tf.
		SetEnvVars(generateTerraformInfraVariablesEnvironment(infrastructure.Spec.SecretRef)...).
		Destroy(ctx)
}

func (a *actuator) ForceDelete(_ context.Context, _ logr.Logger, _ *extensionsv1alpha1.Infrastructure, _ *extensionscontroller.Cluster) error {
	return nil
}
