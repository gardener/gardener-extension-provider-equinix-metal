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
	"path/filepath"

	"github.com/gardener/gardener-extension-provider-packet/pkg/packet"
	extensionscontroller "github.com/gardener/gardener/extensions/pkg/controller"
	"github.com/gardener/gardener/extensions/pkg/controller/controlplane/genericactuator"
	v1beta1constants "github.com/gardener/gardener/pkg/apis/core/v1beta1/constants"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/gardener/gardener/pkg/utils/chart"
	kutil "github.com/gardener/gardener/pkg/utils/kubernetes"
	"github.com/gardener/gardener/pkg/utils/secrets"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apiserver/pkg/authentication/user"
)

var controlPlaneSecrets = &secrets.Secrets{
	CertificateSecretConfigs: map[string]*secrets.CertificateSecretConfig{
		v1beta1constants.SecretNameCACluster: {
			Name:       v1beta1constants.SecretNameCACluster,
			CommonName: "kubernetes",
			CertType:   secrets.CACert,
		},
	},
	SecretConfigsFunc: func(cas map[string]*secrets.Certificate, clusterName string) []secrets.ConfigInterface {
		return []secrets.ConfigInterface{
			&secrets.ControlPlaneSecretConfig{
				CertificateSecretConfig: &secrets.CertificateSecretConfig{
					Name:         packet.CloudControllerManagerImageName,
					CommonName:   "system:cloud-controller-manager",
					Organization: []string{user.SystemPrivilegedGroup},
					CertType:     secrets.ClientCert,
					SigningCA:    cas[v1beta1constants.SecretNameCACluster],
				},
				KubeConfigRequest: &secrets.KubeConfigRequest{
					ClusterName:  clusterName,
					APIServerURL: v1beta1constants.DeploymentNameKubeAPIServer,
				},
			},
		}
	},
}

var controlPlaneChart = &chart.Chart{
	Name: "seed-controlplane",
	Path: filepath.Join(packet.InternalChartsPath, "seed-controlplane"),
	SubCharts: []*chart.Chart{
		{
			Name:   "cloud-provider-equinix-metal",
			Images: []string{packet.CloudControllerManagerImageName},
			Objects: []*chart.Object{
				{Type: &corev1.Service{}, Name: "cloud-controller-manager"},
				{Type: &appsv1.Deployment{}, Name: "cloud-controller-manager"},
				{Type: &corev1.ConfigMap{}, Name: "cloud-controller-manager-observability-config"},
			},
		},
	},
}

var controlPlaneShootChart = &chart.Chart{
	Name: "shoot-system-components",
	Path: filepath.Join(packet.InternalChartsPath, "shoot-system-components"),
	SubCharts: []*chart.Chart{
		{
			Name: "cloud-provider-equinix-metal",
			Objects: []*chart.Object{
				{Type: &rbacv1.ClusterRole{}, Name: "system:controller:cloud-node-controller"},
				{Type: &rbacv1.ClusterRoleBinding{}, Name: "system:controller:cloud-node-controller"},
			},
		},
		{
			Name:    "metallb",
			Images:  []string{packet.MetalLBControllerImageName, packet.MetalLBSpeakerImageName},
			Objects: []*chart.Object{},
		},
	},
}

var storageClassChart = &chart.Chart{
	Name: "shoot-storageclasses",
	Path: filepath.Join(packet.InternalChartsPath, "shoot-storageclasses"),
}

// NewValuesProvider creates a new ValuesProvider for the generic actuator.
func NewValuesProvider(logger logr.Logger) genericactuator.ValuesProvider {
	return &valuesProvider{
		logger: logger.WithName("packet-values-provider"),
	}
}

// valuesProvider is a ValuesProvider that provides Packet-specific values for the 2 charts applied by the generic actuator.
type valuesProvider struct {
	genericactuator.NoopValuesProvider
	logger logr.Logger
}

// GetControlPlaneChartValues returns the values for the control plane chart applied by the generic actuator.
func (vp *valuesProvider) GetControlPlaneChartValues(
	ctx context.Context,
	cp *extensionsv1alpha1.ControlPlane,
	cluster *extensionscontroller.Cluster,
	checksums map[string]string,
	scaledDown bool,
) (map[string]interface{}, error) {
	// TODO: Remove this code in next version. Delete old config
	if err := vp.deleteCCMMonitoringConfig(ctx, cp.Namespace); err != nil {
		return nil, err
	}
	// Get control plane chart values
	return getControlPlaneChartValues(cp, cluster, checksums, scaledDown)
}

// GetControlPlaneShootChartValues returns the values for the control plane shoot chart applied by the generic actuator.
func (vp *valuesProvider) GetControlPlaneShootChartValues(
	ctx context.Context,
	cp *extensionsv1alpha1.ControlPlane,
	cluster *extensionscontroller.Cluster,
	checksum map[string]string,
) (map[string]interface{}, error) {
	// Get credentials from the referenced secret
	credentials, err := vp.getCredentials(ctx, cp)
	if err != nil {
		return nil, err
	}

	// Get control plane shoot chart values
	return getControlPlaneShootChartValues(cluster, credentials)
}

// getCredentials determines the credentials from the secret referenced in the ControlPlane resource.
func (vp *valuesProvider) getCredentials(
	ctx context.Context,
	cp *extensionsv1alpha1.ControlPlane,
) (*packet.Credentials, error) {
	secret, err := extensionscontroller.GetSecretByReference(ctx, vp.Client(), &cp.Spec.SecretRef)
	if err != nil {
		return nil, errors.Wrapf(err, "could not get secret by reference for controlplane '%s'", kutil.ObjectName(cp))
	}
	credentials, err := packet.ReadCredentialsSecret(secret)
	if err != nil {
		return nil, errors.Wrapf(err, "could not read credentials from secret '%s'", kutil.ObjectName(secret))
	}
	return credentials, nil
}

// getCCMChartValues collects and returns the CCM chart values.
func getControlPlaneChartValues(
	cp *extensionsv1alpha1.ControlPlane,
	cluster *extensionscontroller.Cluster,
	checksums map[string]string,
	scaledDown bool,
) (map[string]interface{}, error) {
	values := map[string]interface{}{
		"cloud-provider-equinix-metal": map[string]interface{}{
			"replicas":          extensionscontroller.GetControlPlaneReplicas(cluster, scaledDown, 1),
			"clusterName":       cp.Namespace,
			"kubernetesVersion": cluster.Shoot.Spec.Kubernetes.Version,
			"podNetwork":        extensionscontroller.GetPodNetwork(cluster),
			"podAnnotations": map[string]interface{}{
				"checksum/secret-cloud-controller-manager": checksums[packet.CloudControllerManagerImageName],
				"checksum/secret-cloudprovider":            checksums[v1beta1constants.SecretNameCloudProvider],
			},
			"facility": cluster.Shoot.Spec.Region,
		},
		"metallb": map[string]interface{}{},
	}

	return values, nil
}

// getControlPlaneShootChartValues collects and returns the control plane shoot chart values.
func getControlPlaneShootChartValues(
	cluster *extensionscontroller.Cluster,
	credentials *packet.Credentials,
) (map[string]interface{}, error) {

	return map[string]interface{}{}, nil
}
