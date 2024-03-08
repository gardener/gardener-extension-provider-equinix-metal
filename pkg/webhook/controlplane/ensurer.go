// Copyright (c) 2019 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package controlplane

import (
	"context"
	"fmt"

	"github.com/Masterminds/semver/v3"
	"github.com/coreos/go-systemd/v22/unit"
	extensionswebhook "github.com/gardener/gardener/extensions/pkg/webhook"
	gcontext "github.com/gardener/gardener/extensions/pkg/webhook/context"
	"github.com/gardener/gardener/extensions/pkg/webhook/controlplane"
	"github.com/gardener/gardener/extensions/pkg/webhook/controlplane/genericmutator"
	v1beta1constants "github.com/gardener/gardener/pkg/apis/core/v1beta1/constants"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/gardener/gardener/pkg/component/machinecontrollermanager"
	kutil "github.com/gardener/gardener/pkg/utils/kubernetes"
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	vpaautoscalingv1 "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	kubeletconfigv1beta1 "k8s.io/kubelet/config/v1beta1"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/gardener/gardener-extension-provider-equinix-metal/imagevector"
	eqxcontrolplane "github.com/gardener/gardener-extension-provider-equinix-metal/pkg/controller/controlplane"
	"github.com/gardener/gardener-extension-provider-equinix-metal/pkg/equinixmetal"
)

// NewEnsurer creates a new controlplane ensurer.
func NewEnsurer(client client.Client, logger logr.Logger, gardenletManagesMCM bool) genericmutator.Ensurer {
	return &ensurer{
		logger:              logger.WithName("equinix-metal-controlplane-ensurer"),
		client:              client,
		gardenletManagesMCM: gardenletManagesMCM,
	}
}

type ensurer struct {
	genericmutator.NoopEnsurer
	client              client.Client
	logger              logr.Logger
	gardenletManagesMCM bool
}

// ImageVector is exposed for testing.
var ImageVector = imagevector.ImageVector()

// EnsureMachineControllerManagerDeployment ensures that the machine-controller-manager deployment conforms to the provider requirements.
func (e *ensurer) EnsureMachineControllerManagerDeployment(_ context.Context, _ gcontext.GardenContext, newObj, _ *appsv1.Deployment) error {
	if !e.gardenletManagesMCM {
		return nil
	}

	image, err := ImageVector.FindImage(equinixmetal.MachineControllerManagerEquinixMetalImageName)
	if err != nil {
		return err
	}

	newObj.Spec.Template.Spec.Containers = extensionswebhook.EnsureContainerWithName(
		newObj.Spec.Template.Spec.Containers,
		machinecontrollermanager.ProviderSidecarContainer(newObj.Namespace, equinixmetal.Name, image.String()),
	)
	return nil
}

// EnsureMachineControllerManagerVPA ensures that the machine-controller-manager VPA conforms to the provider requirements.
func (e *ensurer) EnsureMachineControllerManagerVPA(_ context.Context, _ gcontext.GardenContext, newObj, _ *vpaautoscalingv1.VerticalPodAutoscaler) error {
	if !e.gardenletManagesMCM {
		return nil
	}

	var (
		minAllowed = corev1.ResourceList{}
		maxAllowed = corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("2"),
			corev1.ResourceMemory: resource.MustParse("5G"),
		}
	)

	if newObj.Spec.ResourcePolicy == nil {
		newObj.Spec.ResourcePolicy = &vpaautoscalingv1.PodResourcePolicy{}
	}

	newObj.Spec.ResourcePolicy.ContainerPolicies = extensionswebhook.EnsureVPAContainerResourcePolicyWithName(
		newObj.Spec.ResourcePolicy.ContainerPolicies,
		machinecontrollermanager.ProviderSidecarVPAContainerPolicy(equinixmetal.Name, minAllowed, maxAllowed),
	)
	return nil
}

// EnsureKubeAPIServerDeployment ensures that the kube-apiserver deployment conforms to the provider requirements.
func (e *ensurer) EnsureKubeAPIServerDeployment(ctx context.Context, gctx gcontext.GardenContext, new, old *appsv1.Deployment) error {
	cluster, err := gctx.GetCluster(ctx)
	if err != nil {
		return err
	}
	ps := &new.Spec.Template.Spec
	if c := extensionswebhook.ContainerWithName(ps.Containers, "kube-apiserver"); c != nil {
		if err := ensureKubeAPIServerCommandLineArgs(c, cluster.Shoot.Spec.Kubernetes.Version); err != nil {
			return err
		}
	}

	return controlplane.EnsureSecretChecksumAnnotation(ctx, &new.Spec.Template, e.client, new.Namespace, v1beta1constants.SecretNameCloudProvider)
}

// EnsureKubeControllerManagerDeployment ensures that the kube-controller-manager deployment conforms to the provider requirements.
func (e *ensurer) EnsureKubeControllerManagerDeployment(ctx context.Context, gctx gcontext.GardenContext, new, old *appsv1.Deployment) error {
	ps := &new.Spec.Template.Spec
	if c := extensionswebhook.ContainerWithName(ps.Containers, "kube-controller-manager"); c != nil {
		ensureKubeControllerManagerCommandLineArgs(c)
	}
	return nil
}

// EnsureVPNSeedServerDeployment ensures that the vpn-seed-server deployment conforms to the provider requirements.
func (e *ensurer) EnsureVPNSeedServerDeployment(ctx context.Context, gCtx gcontext.GardenContext, new, old *appsv1.Deployment) error {
	cluster, err := gCtx.GetCluster(ctx)
	if err != nil {
		return err
	}
	infra := &extensionsv1alpha1.Infrastructure{}
	if err := e.client.Get(ctx, kutil.Key(new.Namespace, cluster.Shoot.Name), infra); err != nil {
		return fmt.Errorf("failed to get %s infrastructure: %v", cluster.Shoot.Name, err)
	}
	if infra.Status.NodesCIDR == nil {
		e.logger.V(2).Info("node cidr not defined")
		return nil
	}
	return eqxcontrolplane.EnsureNodeNetworkOfVpnSeed(
		ctx,
		e.client,
		new.Namespace,
		eqxcontrolplane.ParseJoinedNetwork(*infra.Status.NodesCIDR))
}

// EnsureAdditionalUnits ensures that additional required system units are added.
func (e *ensurer) EnsureAdditionalUnits(ctx context.Context, gctx gcontext.GardenContext, new, _ *[]extensionsv1alpha1.Unit) error {
	var (
		command      = "start"
		trueVar      = true
		bgpRouteUnit = `[Unit]
Description=Routes to BGP peers
After=network.target
Wants=network.target
[Install]
WantedBy=kubelet.service
[Service]
Type=oneshot
RemainAfterExit=yes
ExecStart=/opt/bin/bgp-peer.sh
`
	)

	extensionswebhook.AppendUniqueUnit(new, extensionsv1alpha1.Unit{
		Name:    "bgp-peer-route.service",
		Enable:  &trueVar,
		Command: &command,
		Content: &bgpRouteUnit,
	})
	return nil
}

// EnsureAdditionalFiles ensures that additional required system files are added.
func (e *ensurer) EnsureAdditionalFiles(ctx context.Context, gctx gcontext.GardenContext, new, _ *[]extensionsv1alpha1.File) error {
	var (
		permissions       int32 = 0755
		customFileContent       = `#!/bin/sh
# get my private IP
GATEWAY="$(curl https://metadata.platformequinix.com/metadata | jq -r '.network.addresses[] | select( .address_family == 4 and .public == false ) | .gateway')"
ip route add 169.254.255.1 via ${GATEWAY} dev bond0
ip route add 169.254.255.2 via ${GATEWAY} dev bond0
`
	)

	appendUniqueFile(new, extensionsv1alpha1.File{
		Path:        "/opt/bin/bgp-peer.sh",
		Permissions: &permissions,
		Content: extensionsv1alpha1.FileContent{
			Inline: &extensionsv1alpha1.FileContentInline{
				Encoding: "",
				Data:     customFileContent,
			},
		},
	})
	return nil
}

// appendUniqueFile appends a unit file only if it does not exist, otherwise overwrite content of previous files
func appendUniqueFile(files *[]extensionsv1alpha1.File, file extensionsv1alpha1.File) {
	resFiles := make([]extensionsv1alpha1.File, 0, len(*files))

	for _, f := range *files {
		if f.Path != file.Path {
			resFiles = append(resFiles, f)
		}
	}

	*files = append(resFiles, file)
}

func ensureKubeAPIServerCommandLineArgs(c *corev1.Container, k8sVersion string) error {
	// Ensure CSI-related admission plugins
	c.Command = extensionswebhook.EnsureNoStringWithPrefixContains(c.Command, "--enable-admission-plugins=",
		"PersistentVolumeLabel", ",")
	c.Command = extensionswebhook.EnsureStringWithPrefixContains(c.Command, "--disable-admission-plugins=",
		"PersistentVolumeLabel", ",")

	// Ensure CSI-related feature gates
	c.Command = extensionswebhook.EnsureNoStringWithPrefixContains(c.Command, "--feature-gates=",
		"VolumeSnapshotDataSource=false", ",")
	c.Command = extensionswebhook.EnsureNoStringWithPrefixContains(c.Command, "--feature-gates=",
		"CSINodeInfo=false", ",")
	c.Command = extensionswebhook.EnsureNoStringWithPrefixContains(c.Command, "--feature-gates=",
		"CSIDriverRegistry=false", ",")
	c.Command = extensionswebhook.EnsureNoStringWithPrefixContains(c.Command, "--feature-gates=",
		"KubeletPluginsWatcher=false", ",")

	return nil
}

func ensureKubeControllerManagerCommandLineArgs(c *corev1.Container) {
	c.Command = extensionswebhook.EnsureStringWithPrefix(c.Command, "--cloud-provider=", "external")
}

// EnsureKubeletServiceUnitOptions ensures that the kubelet.service unit options conform to the provider requirements.
func (e *ensurer) EnsureKubeletServiceUnitOptions(_ context.Context, _ gcontext.GardenContext, _ *semver.Version, new, _ []*unit.UnitOption) ([]*unit.UnitOption, error) {
	if opt := extensionswebhook.UnitOptionWithSectionAndName(new, "Service", "ExecStart"); opt != nil {
		command := extensionswebhook.DeserializeCommandLine(opt.Value)
		command = ensureKubeletCommandLineArgs(command)
		opt.Value = extensionswebhook.SerializeCommandLine(command, 1, " \\\n    ")
	}
	return new, nil
}

func ensureKubeletCommandLineArgs(command []string) []string {
	return extensionswebhook.EnsureStringWithPrefix(command, "--cloud-provider=", "external")
}

// EnsureKubeletConfiguration ensures that the kubelet configuration conforms to the provider requirements.
func (e *ensurer) EnsureKubeletConfiguration(_ context.Context, _ gcontext.GardenContext, _ *semver.Version, new, _ *kubeletconfigv1beta1.KubeletConfiguration) error {
	new.EnableControllerAttachDetach = pointer.Bool(true)

	return nil
}
