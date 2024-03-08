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

	extensionsconfig "github.com/gardener/gardener/extensions/pkg/apis/config"
	"github.com/gardener/gardener/extensions/pkg/util"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	kutil "github.com/gardener/gardener/pkg/utils/kubernetes"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/gardener/gardener-extension-provider-equinix-metal/pkg/controller/controlplane"
	"github.com/gardener/gardener-extension-provider-equinix-metal/pkg/equinixmetal"
	eqxcmclient "github.com/gardener/gardener-extension-provider-equinix-metal/pkg/equinixmetal/client"
)

func (w *workerDelegate) PostReconcileHook(ctx context.Context) error {
	const (
		equinixMetalPrivateNetworkAnnotations = "metal.equinix.com/network-4-private"
	)

	// get the private IPs and providerIDs from the shoot nodes
	_, shootClient, err := util.NewClientForShoot(ctx, w.client, w.worker.Namespace, client.Options{}, extensionsconfig.RESTOptions{})
	if err != nil {
		return err
	}

	credentials, err := equinixmetal.GetCredentialsFromSecretRef(ctx, w.client, w.worker.Spec.SecretRef)
	if err != nil {
		return fmt.Errorf("could not get credentials from secret: %v", err)
	}

	equinixClient, err := eqxcmclient.NewClient(string(credentials.APIToken))
	if err != nil {
		return err
	}

	shootNodes := &corev1.NodeList{}
	if err := shootClient.List(ctx, shootNodes); err != nil {
		return fmt.Errorf("failed to get shoot nodes: %v", err)
	}

	// go through each node, for each one without the right annotation, get the private network
	targetCIDRs := sets.New[string]()
	for _, n := range shootNodes.Items {
		if n.Annotations[equinixMetalPrivateNetworkAnnotations] == "" {
			// we didn't have it, so get it from the Equinix Metal API, and save it
			deviceID, err := deviceIDFromProviderID(n.Spec.ProviderID)
			if deviceID == "" || err != nil {
				continue
			}

			nodePrivateNetwork, err := GetNodePrivateNetwork(ctx, equinixClient, deviceID)
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

	infra := &extensionsv1alpha1.Infrastructure{}
	if err := w.client.Get(ctx, kutil.Key(w.worker.Namespace, w.worker.Name), infra); err != nil {
		return fmt.Errorf("failed to get %s infrastructure: %v", w.worker.Name, err)
	}

	if infra.Status.NodesCIDR == nil ||
		controlplane.ParseJoinedNetwork(*infra.Status.NodesCIDR).Equal(targetCIDRs) {

		var (
			patch         = client.StrategicMergeFrom(infra.DeepCopy())
			joinedNetwork = controlplane.JoinedNetworksCidr(targetCIDRs)
		)

		infra.Status.NodesCIDR = &joinedNetwork
		if err := w.client.Patch(ctx, infra, patch); err != nil {
			return err
		}
	}

	return controlplane.EnsureNodeNetworkOfVpnSeed(ctx, w.client, w.worker.Namespace, targetCIDRs)
}

// PreReconcileHook implements genericactuator.WorkerDelegate.
func (w *workerDelegate) PreReconcileHook(_ context.Context) error {
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

// GetNodePrivateNetwork use the Equinix Metal API to get the CIDR of the private network given a providerID.
func GetNodePrivateNetwork(ctx context.Context, equinixClient eqxcmclient.ClientInterface, deviceID string) (string, error) {
	device, err := equinixClient.GetDevice(ctx, deviceID)
	if err != nil {
		return "", err
	}

	for _, net := range device.IpAddresses {
		// we only want the private, management, ipv4 network
		if net.GetPublic() || !net.GetManagement() || net.GetAddressFamily() != 4 {
			continue
		}

		parent := net.ParentBlock
		if parent == nil || parent.GetNetwork() == "" || parent.GetCidr() == 0 {
			return "", fmt.Errorf("no network information provided for private address %s", net.GetNetwork())
		}

		return fmt.Sprintf("%s/%d", parent.GetNetwork(), parent.GetCidr()), nil
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
