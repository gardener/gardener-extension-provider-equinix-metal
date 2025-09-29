// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

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
	"github.com/gardener/gardener/pkg/component/nodemanagement/machinecontrollermanager"
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	vpaautoscalingv1 "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	kubeletconfigv1beta1 "k8s.io/kubelet/config/v1beta1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/gardener/gardener-extension-provider-equinix-metal/imagevector"
	eqxcontrolplane "github.com/gardener/gardener-extension-provider-equinix-metal/pkg/controller/controlplane"
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
func (e *ensurer) EnsureMachineControllerManagerDeployment(ctx context.Context, gctx gcontext.GardenContext, newObj, _ *appsv1.Deployment) error {
	image, err := ImageVector.FindImage(equinixmetal.MachineControllerManagerEquinixMetalImageName)
	if err != nil {
		return err
	}

	cluster, err := gctx.GetCluster(ctx)
	if err != nil {
		return fmt.Errorf("failed reading Cluster: %w", err)
	}

	newObj.Spec.Template.Spec.Containers = extensionswebhook.EnsureContainerWithName(
		newObj.Spec.Template.Spec.Containers,
		machinecontrollermanager.ProviderSidecarContainer(cluster.Shoot, newObj.Namespace, equinixmetal.Name, image.String()),
	)
	return nil
}

// EnsureMachineControllerManagerVPA ensures that the machine-controller-manager VPA conforms to the provider requirements.
func (e *ensurer) EnsureMachineControllerManagerVPA(_ context.Context, _ gcontext.GardenContext, newObj, _ *vpaautoscalingv1.VerticalPodAutoscaler) error {

	if newObj.Spec.ResourcePolicy == nil {
		newObj.Spec.ResourcePolicy = &vpaautoscalingv1.PodResourcePolicy{}
	}

	newObj.Spec.ResourcePolicy.ContainerPolicies = extensionswebhook.EnsureVPAContainerResourcePolicyWithName(
		newObj.Spec.ResourcePolicy.ContainerPolicies,
		machinecontrollermanager.ProviderSidecarVPAContainerPolicy(equinixmetal.Name),
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
	if err := e.client.Get(ctx, client.ObjectKey{Namespace: new.Namespace,
		Name: cluster.Shoot.Name}, infra); err != nil {
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

// EnsureVPNSeedServerStatefulset ensures that the vpn-seed-server statefulset conforms to the provider requirements.
func (e *ensurer) EnsureVPNSeedServerStatefulSet(ctx context.Context, gCtx gcontext.GardenContext, new, old *appsv1.StatefulSet) error {
	cluster, err := gCtx.GetCluster(ctx)
	if err != nil {
		return err
	}
	infra := &extensionsv1alpha1.Infrastructure{}
	if err := e.client.Get(ctx, client.ObjectKey{Namespace: new.Namespace,
		Name: cluster.Shoot.Name}, infra); err != nil {
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

// EnsureAdditionalProvisionUnits ensures that additional required system units are added, that are required during provisioning.
func (e *ensurer) EnsureAdditionalProvisionUnits(ctx context.Context, gctx gcontext.GardenContext, new, _ *[]extensionsv1alpha1.Unit) error {

	// Define LVM setup unit
	var lvmSetup = `
[Unit]
        Description=LVM Setup
        DefaultDependencies=no
        Before=local-fs-pre.target
        [Service]
        Type=oneshot
        Restart=on-failure
        RemainAfterExit=yes
        ExecStart=/opt/bin/lvm.sh
        [Install]
        WantedBy=multi-user.target
`
	// Define LVM mount unit
	var lvmMountContainerd = `
[Unit]
        Description=Mount LVM to containerd dir
        After=lvm-setup.service
        Before=containerd.service
        [Mount]
        What=/dev/vg-containerd/vol_containerd
        Where=/var/lib/containerd
        Type=ext4
        Options=defaults
        [Install]
        WantedBy=local-fs.target
`
	volume, err := hasVolume(ctx, gctx)
	if err != nil {
		return err
	}
	if !volume {
		return nil
	}
	operatingsystems, err := getOperatingSystems(ctx, gctx)
	if err != nil {
		return err
	}
	if len(operatingsystems) > 1 {
		return fmt.Errorf("multiple operatingsystems used: %v; Only one allowed", operatingsystems)
	}

	switch operatingsystems[0] {
	case "flatcar":
		extensionswebhook.AppendUniqueUnit(new, extensionsv1alpha1.Unit{
			Name:    "lvm-setup.service",
			Enable:  ptr.To(true),
			Command: ptr.To(extensionsv1alpha1.CommandStart),
			Content: ptr.To(lvmSetup),
		})

		extensionswebhook.AppendUniqueUnit(new, extensionsv1alpha1.Unit{
			Name:    "var-lib-containerd.mount",
			Enable:  ptr.To(true),
			Command: ptr.To(extensionsv1alpha1.CommandStart),
			Content: ptr.To(lvmMountContainerd),
		})

		return nil
	}

	return nil
}

// EnsureAdditionalProvisionFiles ensures that additional required system files are added, that are required during provisioning.
func (e *ensurer) EnsureAdditionalProvisionFiles(ctx context.Context, gctx gcontext.GardenContext, new, _ *[]extensionsv1alpha1.File) error {
	volume, err := hasVolume(ctx, gctx)
	if err != nil {
		return err
	}
	if volume {
		operatingsystems, err := getOperatingSystems(ctx, gctx)
		if err != nil {
			return err
		}
		if len(operatingsystems) > 1 {
			return fmt.Errorf("multiple operatingsystems used: %v; Only one allowed", operatingsystems)
		}

		switch operatingsystems[0] {
		case "flatcar":

			var (
				permissions       uint32 = 0755
				customFileContent        = `#!/bin/bash
          set -euo pipefail


          # Function to find all disks
          find_volumes(){
            lsblk -d -o NAME,TYPE | awk '$2 == "disk" {print "/dev/" $1}'
          }

          create_pvs() {
            disks=$(find_volumes)
            usable_disks=""

            for disk in $disks; do
              if pvcreate -y $disk 2>&1 | grep -q "successfully created"; then
                usable_disks="$usable_disks $disk"
              fi
            done

            echo $usable_disks
          }

          # Get the list of usable disks
          usable_disks=$(create_pvs)

          # Create Physical Volumes
          pvcreate ${usable_disks}

          # Create Volume Group
          vgcreate vg-containerd ${usable_disks}

          # Create Logical Volume for data
          lvcreate -n vol_containerd -l 100%FREE vg-containerd

          # Format the data volume with ext4 filesystem
          mkfs.ext4 /dev/vg-containerd/vol_containerd
`
			)

			appendUniqueFile(new, extensionsv1alpha1.File{
				Path:        "/opt/bin/lvm.sh",
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
	}
	return nil
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
		permissions       uint32 = 0755
		customFileContent        = `#!/bin/sh
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
func (e *ensurer) EnsureKubeletServiceUnitOptions(ctx context.Context, gctx gcontext.GardenContext, _ *semver.Version, new, _ []*unit.UnitOption) ([]*unit.UnitOption, error) {
	if opt := extensionswebhook.UnitOptionWithSectionAndName(new, "Service", "ExecStart"); opt != nil {
		command := extensionswebhook.DeserializeCommandLine(opt.Value)
		command = ensureKubeletCommandLineArgs(command)

		volume, err := hasVolume(ctx, gctx)
		if err != nil {
			return new, err
		}
		if volume {
			command = ensureKubeletRootDirCommandLineArg(command)
		}
		opt.Value = extensionswebhook.SerializeCommandLine(command, 1, " \\\n    ")
	}
	return new, nil
}

// getOperatingSystems returns an array of all operatingsytems used in the given cluster
func getOperatingSystems(ctx context.Context, gctx gcontext.GardenContext) ([]string, error) {
	operatingsystems := make(map[string]int)

	cluster, err := gctx.GetCluster(ctx)
	if err != nil {
		return []string{}, err
	}

	for _, worker := range cluster.Shoot.Spec.Provider.Workers {
		operatingsystems[worker.Machine.Image.Name]++
	}
	return getMapKeys(operatingsystems), nil
}

// getMapKeys takes a map[string](int) and returns the keys as []string
func getMapKeys(inputMap map[string]int) []string {
	var output []string
	for k := range inputMap {
		output = append(output, k)
	}
	return output
}

// hasVolume checks if any worker has set a value for `Volume`
func hasVolume(ctx context.Context, gctx gcontext.GardenContext) (bool, error) {
	cluster, err := gctx.GetCluster(ctx)
	if err != nil {
		return false, err
	}
	for _, worker := range cluster.Shoot.Spec.Provider.Workers {
		if worker.Volume != nil {
			return true, nil
		}
	}
	return false, nil
}

// ensureKubeletRootDirCommandLineArg adds a flag to the kubelet to use /var/lib/containerd which is what where we also mount the created LVM if `volume` is set in the worker config
func ensureKubeletRootDirCommandLineArg(command []string) []string {
	return extensionswebhook.EnsureStringWithPrefix(command, "--root-dir=", "/var/lib/containerd")
}

func ensureKubeletCommandLineArgs(command []string) []string {
	return extensionswebhook.EnsureStringWithPrefix(command, "--cloud-provider=", "external")
}

// EnsureKubeletConfiguration ensures that the kubelet configuration conforms to the provider requirements.
func (e *ensurer) EnsureKubeletConfiguration(_ context.Context, _ gcontext.GardenContext, _ *semver.Version, new, _ *kubeletconfigv1beta1.KubeletConfiguration) error {
	new.EnableControllerAttachDetach = ptr.To(true)

	return nil
}
