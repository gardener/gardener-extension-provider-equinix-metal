// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package shoot

import (
	"context"

	extensionswebhook "github.com/gardener/gardener/extensions/pkg/webhook"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/meta"
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

func (m *mutator) Mutate(ctx context.Context, new, old client.Object) error {
	acc, err := meta.Accessor(new)
	if err != nil {
		return errors.Wrapf(err, "could not create accessor during webhook")
	}
	// If the object does have a deletion timestamp then we don't want to mutate anything.
	if acc.GetDeletionTimestamp() != nil {
		return nil
	}

	switch x := new.(type) {
	case *appsv1.Deployment:
		switch x.Name {
		case "metrics-server":
			extensionswebhook.LogMutation(logger, x.Kind, x.Namespace, x.Name)
			return m.mutateMetricsServerDeployment(ctx, x)
		case "vpn-shoot":
			extensionswebhook.LogMutation(logger, x.Kind, x.Namespace, x.Name)
			return m.mutateVPNShootDeployment(ctx, x)
		}
	}
	return nil
}
