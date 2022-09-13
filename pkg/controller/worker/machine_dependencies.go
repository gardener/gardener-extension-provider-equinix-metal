// Copyright (c) 2020 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
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

package worker

import (
	"context"
	"fmt"
	"strings"

	"github.com/gardener/gardener-extension-provider-equinix-metal/pkg/equinixmetal"
	eqxcmclient "github.com/gardener/gardener-extension-provider-equinix-metal/pkg/equinixmetal/client"
	extensionsconfig "github.com/gardener/gardener/extensions/pkg/apis/config"

	"github.com/gardener/gardener/extensions/pkg/util"
	v1beta1constants "github.com/gardener/gardener/pkg/apis/core/v1beta1/constants"
	kutil "github.com/gardener/gardener/pkg/utils/kubernetes"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (w *workerDelegate) DeployMachineDependencies(_ context.Context) error {
	return nil
}

func (w *workerDelegate) CleanupMachineDependencies(ctx context.Context) error {
	const (
		nodeNetworkEnvVarKey                  = "NODE_NETWORK"
		equinixMetalPrivateNetworkAnnotations = "metal.equinix.com/network-4-private"
	)

	// get the private IPs and providerIDs from the shoot nodes
	_, shootClient, err := util.NewClientForShoot(ctx, w.Client(), w.worker.Namespace, client.Options{}, extensionsconfig.RESTOptions{})
	if err != nil {
		return err
	}

	shootNodes := &corev1.NodeList{}
	if err := shootClient.List(ctx, shootNodes); err != nil {
		return fmt.Errorf("failed to get shoot nodes: %v", err)
	}

	// go through each node, for each one without the right annotation, get the private network
	targetCIDRs := sets.NewString()
	for _, n := range shootNodes.Items {
		if n.Annotations[equinixMetalPrivateNetworkAnnotations] == "" {
			// we didn't have it, so get it from the Equinix Metal API, and save it
			deviceID, err := deviceIDFromProviderID(n.Spec.ProviderID)
			if deviceID == "" || err != nil {
				continue
			}

			nodePrivateNetwork, err := getNodePrivateNetwork(ctx, deviceID, w.Client(), w.worker.Spec.SecretRef)
			if err != nil {
				return fmt.Errorf("error getting private network from Equinix Metal API for %s: %v", n.Spec.ProviderID, err)
			}

			if nodePrivateNetwork == "" {
				continue
			}

			if n.Annotations[equinixMetalPrivateNetworkAnnotations] == nodePrivateNetwork {
				continue
			}

			// if it was not set already, set it and save it
			patch := client.StrategicMergeFrom(n.DeepCopy())
			metav1.SetMetaDataAnnotation(&n.ObjectMeta, equinixMetalPrivateNetworkAnnotations, nodePrivateNetwork)
			if err := shootClient.Patch(ctx, &n, patch); err != nil {
				return fmt.Errorf("unable to patch node %s with private network cidr: %v", n.Name, err)
			}
		}

		targetCIDRs.Insert(n.Annotations[equinixMetalPrivateNetworkAnnotations])
	}

	// Check if the `vpn-seed-server` deployment exists. If yes then the ReversedVPN feature gate is enabled in
	// gardenlet and we have to configure the `vpn-seed` container here. Otherwise, the ReversedVPN feature gate is
	// disabled and the `vpn-seed` container resides in the `kube-apiserver` deployment.
	var (
		vpnSeedContainerName = "vpn-seed-server"
		deploy               = &appsv1.Deployment{}
	)

	if err := w.Client().Get(ctx, kutil.Key(w.worker.Namespace, v1beta1constants.DeploymentNameVPNSeedServer), deploy); err != nil {
		if !apierrors.IsNotFound(err) {
			return fmt.Errorf("failed to get %s deployment: %v", v1beta1constants.DeploymentNameVPNSeedServer, err)
		}

		if err2 := w.Client().Get(ctx, kutil.Key(w.worker.Namespace, v1beta1constants.DeploymentNameKubeAPIServer), deploy); err2 != nil {
			return fmt.Errorf("failed to get %s deployment: %v", v1beta1constants.DeploymentNameKubeAPIServer, err2)
		}
		vpnSeedContainerName = "vpn-seed"
	}

	var (
		envVarExists  bool
		envVarChanged bool

		patch                  = client.StrategicMergeFrom(deploy.DeepCopy())
		nodeNetworkEnvVarValue = strings.Join(targetCIDRs.List(), ",")
	)

	for i, ctr := range deploy.Spec.Template.Spec.Containers {
		if ctr.Name != vpnSeedContainerName {
			continue
		}

		for j, env := range ctr.Env {
			if env.Name != nodeNetworkEnvVarKey {
				continue
			}

			envVarExists = true

			if env.Value != nodeNetworkEnvVarValue {
				deploy.Spec.Template.Spec.Containers[i].Env[j].Value = nodeNetworkEnvVarValue
				envVarChanged = true
			}
		}

		if !envVarExists {
			deploy.Spec.Template.Spec.Containers[i].Env = append(ctr.Env, corev1.EnvVar{Name: nodeNetworkEnvVarKey, Value: nodeNetworkEnvVarValue})
			envVarChanged = true
		}
	}

	if !envVarChanged {
		return nil
	}

	return w.Client().Patch(ctx, deploy, patch)
}

// PreReconcileHook implements genericactuator.WorkerDelegate.
func (w *workerDelegate) PreReconcileHook(_ context.Context) error {
	return nil
}

// PostReconcileHook implements genericactuator.WorkerDelegate.
func (w *workerDelegate) PostReconcileHook(_ context.Context) error {
	return nil
}

// PreDeleteHook implements genericactuator.WorkerDelegate.
func (w *workerDelegate) PreDeleteHook(_ context.Context) error {
	return nil
}

// PostDeleteHook implements genericactuator.WorkerDelegate.
func (w *workerDelegate) PostDeleteHook(_ context.Context) error {
	return nil
}

// getNodePrivateNetwork use the Equinix Metal API to get the CIDR of the private network given a providerID.
func getNodePrivateNetwork(ctx context.Context, deviceID string, kClient client.Client, secretRef corev1.SecretReference) (string, error) {
	credentials, err := equinixmetal.GetCredentialsFromSecretRef(ctx, kClient, secretRef)
	if err != nil {
		return "", fmt.Errorf("could not get credentials from secret: %v", err)
	}

	pClient := eqxcmclient.NewClient(string(credentials.APIToken))

	device, err := pClient.DeviceGet(deviceID)
	if err != nil {
		return "", err
	}

	for _, net := range device.Network {
		// we only want the private, management, ipv4 network
		if net.Public || !net.Management || net.AddressFamily != 4 {
			continue
		}

		parent := net.ParentBlock
		if parent == nil || parent.Network == "" || parent.CIDR == 0 {
			return "", fmt.Errorf("no network information provided for private address %s", net.String())
		}

		return fmt.Sprintf("%s/%d", parent.Network, parent.CIDR), nil
	}

	return "", nil
}

// deviceIDFromProviderID returns a device's ID from providerID.
//
// The providerID spec should be retrievable from the Kubernetes
// node object. The expected format is: equinixmetal://device-id or just device-id
func deviceIDFromProviderID(providerID string) (string, error) {
	const (
		providerName           = "equinixmetal"
		deprecatedProviderName = "packet"
	)

	if providerID == "" {
		return "", nil
	}

	var (
		deviceID string
		split    = strings.Split(providerID, "://")
	)

	switch len(split) {
	case 2:
		deviceID = split[1]
		if split[0] != providerName && split[0] != deprecatedProviderName {
			return "", nil
		}

	case 1:
		deviceID = providerID

	default:
		return "", errors.Errorf("unexpected providerID format: %s, format should be: 'device-id' or 'equinixmetal://device-id'", providerID)
	}

	return deviceID, nil
}
