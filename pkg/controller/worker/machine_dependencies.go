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
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/gardener/gardener-extension-provider-packet/pkg/packet"
	packetclient "github.com/gardener/gardener-extension-provider-packet/pkg/packet/client"
	util "github.com/gardener/gardener/extensions/pkg/util"
	v1beta1constants "github.com/gardener/gardener/pkg/apis/core/v1beta1/constants"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	vpnSeed                               = "vpn-seed"
	apiServerDeploy                       = v1beta1constants.DeploymentNameKubeAPIServer
	nodeNetworkEnvVar                     = "NODE_NETWORK"
	equinixMetalPrivateNetworkAnnotations = "metal.equinix.com/network-4-private"
	providerName                          = "equinixmetal"
	deprecatedProviderName                = "packet"
)

func (w *workerDelegate) DeployMachineDependencies(_ context.Context) error {
	return nil
}

func (w *workerDelegate) CleanupMachineDependencies(ctx context.Context) error {
	ns := w.worker.GetNamespace()
	c := w.Client()

	// get the private IPs and providerIDs from the shoot nodes
	_, shoot, err := util.NewClientForShoot(ctx, c, ns, client.Options{})
	if err != nil {
		return err
	}

	shootNodes := &corev1.NodeList{}
	if err := shoot.List(ctx, shootNodes); err != nil {
		return fmt.Errorf("failed to get shoot nodes: %v", err)
	}
	// go through each node, for each one without the right annotation, get the private network
	var targetCidrs []string
	for _, n := range shootNodes.Items {
		if n.Annotations == nil || n.Annotations[equinixMetalPrivateNetworkAnnotations] == "" {
			// we didn't have it, so get it from the Equinix Metal API, and save it
			deviceID, err := deviceIDFromProviderID(n.Spec.ProviderID)
			if deviceID == "" || err != nil {
				continue
			}
			nodePrivateNetwork, err := getNodePrivateNetwork(ctx, deviceID, c, ns)
			if err != nil {
				return fmt.Errorf("error getting private network from Equinix Metal API for %s: %v", n.Spec.ProviderID, err)
			}
			if nodePrivateNetwork == "" {
				continue
			}
			if n.Annotations == nil {
				n.Annotations = map[string]string{}
			}
			// if it was not set already, set it and save it
			if n.Annotations[equinixMetalPrivateNetworkAnnotations] != nodePrivateNetwork {
				n.Annotations[equinixMetalPrivateNetworkAnnotations] = nodePrivateNetwork
				// update the node in the k8s cluster ***
				patch, _ := json.Marshal(map[string]interface{}{
					"metadata": map[string]interface{}{
						"annotations": n.Annotations,
					},
				})
				if err := shoot.Patch(ctx, &n, client.RawPatch(types.StrategicMergePatchType, patch)); err != nil {
					return fmt.Errorf("unable to patch node %s with private network cidr: %v", n.Name, err)
				}
			}
		}
		targetCidrs = append(targetCidrs, n.Annotations[equinixMetalPrivateNetworkAnnotations])
	}

	// sort them for consistency; the order really doesn't matter, as long as it is consistent
	targetCidrs = unique(targetCidrs)
	sort.Strings(targetCidrs)
	cidrs := strings.Join(targetCidrs, ",")

	deploy := &appsv1.Deployment{}
	if err := c.Get(ctx, client.ObjectKey{
		Namespace: ns,
		Name:      apiServerDeploy,
	}, deploy); err != nil {
		return fmt.Errorf("failed to get kube-apiserver deployment: %v", err)
	}

	// find the vpn-seed container
	ctrs := deploy.Spec.Template.Spec.Containers
	var changed bool
	for i, ctr := range ctrs {
		// find the right container
		if ctr.Name != vpnSeed {
			continue
		}
		// find the right env var
		var found bool
		for _, env := range ctr.Env {
			if env.Name != nodeNetworkEnvVar {
				continue
			}
			// track that it existed
			found = true
			if env.Value != cidrs {
				env.Value = cidrs
			}
		}
		// if we did not find it, so add it
		if !found {
			ctr.Env = append(ctr.Env, corev1.EnvVar{Name: nodeNetworkEnvVar, Value: cidrs})
		}
		ctrs[i] = ctr
		changed = true
	}
	if !changed {
		return nil
	}
	return c.Update(ctx, deploy)
}

func unique(stringSlice []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range stringSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

// getNodePrivateNetwork use the Equinix Metal API to get the CIDR of the private network given a providerID.
func getNodePrivateNetwork(ctx context.Context, deviceID string, kClient client.Client, namespace string) (string, error) {
	secretRef := corev1.SecretReference{
		Name:      v1beta1constants.SecretNameCloudProvider,
		Namespace: namespace,
	}
	credentials, err := packet.GetCredentialsFromSecretRef(ctx, kClient, secretRef)
	if err != nil {
		return "", fmt.Errorf("could not get credentials from secret: %v", err)
	}
	pClient := packetclient.NewClient(string(credentials.APIToken))

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
	if providerID == "" {
		return "", nil
	}

	split := strings.Split(providerID, "://")
	var deviceID string
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
