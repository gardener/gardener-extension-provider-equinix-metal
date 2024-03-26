// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package shoot

import (
	"context"

	extensionswebhook "github.com/gardener/gardener/extensions/pkg/webhook"
	appsv1 "k8s.io/api/apps/v1"
)

func (m *mutator) mutateMetricsServerDeployment(_ context.Context, deployment *appsv1.Deployment) error {
	if c := extensionswebhook.ContainerWithName(deployment.Spec.Template.Spec.Containers, "metrics-server"); c != nil {
		c.Command = extensionswebhook.EnsureStringWithPrefix(c.Command, "--kubelet-preferred-address-types=", "InternalIP,ExternalIP")
	}

	return nil
}
