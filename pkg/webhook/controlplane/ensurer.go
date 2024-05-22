// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package controlplane

import (
	"context"

	"github.com/Masterminds/semver/v3"
	"github.com/coreos/go-systemd/v22/unit"
	extensionswebhook "github.com/gardener/gardener/extensions/pkg/webhook"
	gcontext "github.com/gardener/gardener/extensions/pkg/webhook/context"
	"github.com/gardener/gardener/extensions/pkg/webhook/controlplane"
	"github.com/gardener/gardener/extensions/pkg/webhook/controlplane/genericmutator"
	v1beta1constants "github.com/gardener/gardener/pkg/apis/core/v1beta1/constants"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/gardener/gardener/pkg/component/nodemanagement/machinecontrollermanager"
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	vpaautoscalingv1 "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	kubeletconfigv1beta1 "k8s.io/kubelet/config/v1beta1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/gardener/gardener-extension-provider-equinix-metal/imagevector"
	"github.com/gardener/gardener-extension-provider-equinix-metal/pkg/equinixmetal"
)

// NewEnsurer creates a new controlplane ensurer.
func NewEnsurer(client client.Client, logger logr.Logger) genericmutator.Ensurer {
	return &ensurer{
		logger: logger.WithName("equinix-metal-controlplane-ensurer"),
		client: client,
	}
}

type ensurer struct {
	genericmutator.NoopEnsurer
	client client.Client
	logger logr.Logger
}

// ImageVector is exposed for testing.
var ImageVector = imagevector.ImageVector()

// EnsureMachineControllerManagerDeployment ensures that the machine-controller-manager deployment conforms to the provider requirements.
func (e *ensurer) EnsureMachineControllerManagerDeployment(_ context.Context, _ gcontext.GardenContext, newObj, _ *appsv1.Deployment) error {
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
	ps := &new.Spec.Template.Spec
	if c := extensionswebhook.ContainerWithName(ps.Containers, "kube-apiserver"); c != nil {
		if err := ensureKubeAPIServerCommandLineArgs(c); err != nil {
			return err
		}
	}

	keepNodeNetworkEnvVarIfPresentInOldDeployment(new, old, "vpn-seed")

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
func (e *ensurer) EnsureVPNSeedServerDeployment(_ context.Context, _ gcontext.GardenContext, new, old *appsv1.Deployment) error {
	keepNodeNetworkEnvVarIfPresentInOldDeployment(new, old, "vpn-seed-server")
	return nil
}

func keepNodeNetworkEnvVarIfPresentInOldDeployment(new, old *appsv1.Deployment, containerName string) {
	if old == nil {
		return
	}

	// Preserve NODE_NETWORK env variable in vpn-seed container. The Worker controller sets this flag at the end of its
	// reconciliation, however, the kube-apiserver deployment is created earlier, so let's keep the value until the
	// Worker controller updates it. This is to not trigger avoidable rollouts/restarts.
	var (
		newContainer        = extensionswebhook.ContainerWithName(new.Spec.Template.Spec.Containers, containerName)
		oldVPNSeedContainer = extensionswebhook.ContainerWithName(old.Spec.Template.Spec.Containers, containerName)

		nodeNetworkEnvVarName  = "NODE_NETWORK"
		nodeNetworkEnvVarValue string
	)

	if oldVPNSeedContainer != nil {
		for _, env := range oldVPNSeedContainer.Env {
			if env.Name == nodeNetworkEnvVarName {
				nodeNetworkEnvVarValue = env.Value
				break
			}
		}
	}

	if newContainer != nil && nodeNetworkEnvVarValue != "" {
		for _, env := range newContainer.Env {
			if env.Name == nodeNetworkEnvVarName && env.Value != nodeNetworkEnvVarValue {
				return
			}
		}

		newContainer.Env = extensionswebhook.EnsureEnvVarWithName(newContainer.Env, corev1.EnvVar{
			Name:  nodeNetworkEnvVarName,
			Value: nodeNetworkEnvVarValue,
		})
	}
}

// EnsureAdditionalUnits ensures that additional required system units are added.
func (e *ensurer) EnsureAdditionalUnits(_ context.Context, _ gcontext.GardenContext, new, _ *[]extensionsv1alpha1.Unit) error {
	extensionswebhook.AppendUniqueUnit(new, extensionsv1alpha1.Unit{
		Name:    "bgp-peer-route.service",
		Enable:  ptr.To(true),
		Command: ptr.To(extensionsv1alpha1.CommandStart),
		Content: ptr.To(`[Unit]
Description=Routes to BGP peers
After=network.target
Wants=network.target
[Install]
WantedBy=kubelet.service
[Service]
Type=oneshot
RemainAfterExit=yes
ExecStart=/opt/bin/bgp-peer.sh
`),
	})
	return nil
}

// EnsureAdditionalFiles ensures that additional required system files are added.
func (e *ensurer) EnsureAdditionalFiles(_ context.Context, _ gcontext.GardenContext, new, _ *[]extensionsv1alpha1.File) error {
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

func ensureKubeAPIServerCommandLineArgs(c *corev1.Container) error {
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
	new.EnableControllerAttachDetach = ptr.To(true)

	return nil
}
