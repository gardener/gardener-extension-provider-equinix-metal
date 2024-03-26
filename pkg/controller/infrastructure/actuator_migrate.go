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

func (a *actuator) Migrate(ctx context.Context, log logr.Logger, infrastructure *extensionsv1alpha1.Infrastructure, cluster *extensionscontroller.Cluster) error {
	log.WithValues("infrastructure", client.ObjectKeyFromObject(infrastructure), "operation", "migrate")
	tf, err := a.newTerraformer(log, equinixmetal.TerraformerPurposeInfra, infrastructure)
	if err != nil {
		return fmt.Errorf("could not create the Terraformer: %+v", err)
	}

	return tf.CleanupConfiguration(ctx)
}
