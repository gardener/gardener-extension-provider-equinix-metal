package controlplane

import (
	"context"
	"strings"

	v1beta1constants "github.com/gardener/gardener/pkg/apis/core/v1beta1/constants"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const nodeNetworkEnvVarKey = "NODE_NETWORK"

func JoinedNetworksCidr(cidrs sets.Set[string]) string {
	return strings.Join(sets.List(cidrs), ",")
}

func ParseJoinedNetwork(joined string) sets.Set[string] {
	nets := strings.Split(joined, ",")
	return sets.New[string](nets...)
}

func EnsureNodeNetworkOfVpnSeed(
	ctx context.Context,
	shootClient client.Client,
	namespace string,
	targetCIDRs sets.Set[string],
) error {
	// Check if the `vpn-seed-server` deployment exists. If yes then the ReversedVPN feature gate is enabled in
	// gardenlet and we have to configure the `vpn-seed` container here. Otherwise, the ReversedVPN feature gate is
	// disabled and the `vpn-seed` container resides in the `kube-apiserver` deployment.
	var (
		vpnSeedContainerName = "vpn-seed-server"
		deploy               = &appsv1.Deployment{}
	)

	// TODO: adjust this to properly check after https://github.com/gardener/gardener/pull/12949 is release
	err := shootClient.Get(ctx, client.ObjectKey{Namespace: namespace, Name: v1beta1constants.DeploymentNameVPNSeedServer}, &appsv1.StatefulSet{})
	if err == nil {
		return nil
	}

	if err = shootClient.Get(ctx, client.ObjectKey{Namespace: namespace, Name: v1beta1constants.DeploymentNameVPNSeedServer}, deploy); err != nil {
		return err
	}

	var (
		envVarExists  bool
		envVarChanged bool

		patch                  = client.StrategicMergeFrom(deploy.DeepCopy())
		nodeNetworkEnvVarValue = JoinedNetworksCidr(targetCIDRs)
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

	return shootClient.Patch(ctx, deploy, patch)
}
