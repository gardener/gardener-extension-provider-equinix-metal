// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package shoot

import (
	"context"
	"path"

	extensionswebhook "github.com/gardener/gardener/extensions/pkg/webhook"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	"github.com/gardener/gardener-extension-provider-equinix-metal/imagevector"
	"github.com/gardener/gardener-extension-provider-equinix-metal/pkg/equinixmetal"
)

const (
	metabotInitContainerName = "metabot"
	metabotVolumeName        = "shared-init-config"
	metabotVolumeMountPath   = "/init-config"
	nodeNetworkFile          = "nodeNetwork"
)

func (m *mutator) mutateVPNShootDeployment(_ context.Context, deployment *appsv1.Deployment) error {
	metabotImage, err := imagevector.ImageVector().FindImage(equinixmetal.MetabotImageName)
	if err != nil {
		return err
	}

	template := &deployment.Spec.Template
	ps := &template.Spec

	for _, initContainer := range ps.InitContainers {
		if initContainer.Name == metabotInitContainerName {
			return nil
		}
	}

	volumeMount := corev1.VolumeMount{
		Name:      metabotVolumeName,
		MountPath: metabotVolumeMountPath,
	}

	ps.InitContainers = append(ps.InitContainers, corev1.Container{
		Name:  metabotInitContainerName,
		Image: metabotImage.String(),
		Args: []string{
			"ip",
			"4",
			"private",
			"parent",
			"network",
			"--out",
			path.Join(metabotVolumeMountPath, nodeNetworkFile),
		},
		VolumeMounts: []corev1.VolumeMount{volumeMount},
	})

	if c := extensionswebhook.ContainerWithName(ps.Containers, "vpn-shoot"); c != nil {
		c.VolumeMounts = extensionswebhook.EnsureVolumeMountWithName(c.VolumeMounts, volumeMount)
	}
	ps.Volumes = extensionswebhook.EnsureVolumeWithName(ps.Volumes, corev1.Volume{
		Name: metabotVolumeName,
		VolumeSource: corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{},
		},
	})

	return nil
}
