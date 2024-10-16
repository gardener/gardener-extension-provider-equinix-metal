// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package shoot

import (
	"context"
	"fmt"

	extensionswebhook "github.com/gardener/gardener/extensions/pkg/webhook"
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type mutator struct {
	logger logr.Logger
}

// NewMutator creates a new Mutator that mutates resources in the shoot cluster.
func NewMutator() extensionswebhook.Mutator {
	return &mutator{
		logger: log.Log.WithName("shoot-mutator"),
	}
}

func (m *mutator) Mutate(ctx context.Context, new, _ client.Object) error {
	deployment, ok := new.(*appsv1.Deployment)
	if !ok {
		return fmt.Errorf("wrong object type %T", new)
	}

	if deployment.GetDeletionTimestamp() != nil {
		return nil
	}

	extensionswebhook.LogMutation(logger, deployment.Kind, deployment.Namespace, deployment.Name)
	return m.mutateVPNShootDeployment(ctx, deployment)
}
