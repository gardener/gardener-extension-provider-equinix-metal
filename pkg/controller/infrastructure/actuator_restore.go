// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package infrastructure

import (
	"context"

	extensionscontroller "github.com/gardener/gardener/extensions/pkg/controller"
	"github.com/gardener/gardener/extensions/pkg/terraformer"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (a *actuator) Restore(ctx context.Context, log logr.Logger, infrastructure *extensionsv1alpha1.Infrastructure, cluster *extensionscontroller.Cluster) error {
	log.WithValues("infrastructure", client.ObjectKeyFromObject(infrastructure), "operation", "restore")
	terraformState, err := terraformer.UnmarshalRawState(infrastructure.Status.State)
	if err != nil {
		return err
	}
	return a.reconcile(ctx, log, infrastructure, cluster, terraformer.CreateOrUpdateState{State: &terraformState.Data})
}
