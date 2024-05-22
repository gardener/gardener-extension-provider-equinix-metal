// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package controlplane

import (
	"context"
	"path/filepath"

	extensionscontroller "github.com/gardener/gardener/extensions/pkg/controller"
	"github.com/gardener/gardener/extensions/pkg/controller/controlplane/genericactuator"
	v1beta1constants "github.com/gardener/gardener/pkg/apis/core/v1beta1/constants"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/gardener/gardener/pkg/utils/chart"
	gutil "github.com/gardener/gardener/pkg/utils/gardener"
	kutil "github.com/gardener/gardener/pkg/utils/kubernetes"
	secretsmanager "github.com/gardener/gardener/pkg/utils/secrets/manager"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/gardener/gardener-extension-provider-equinix-metal/charts"
	api "github.com/gardener/gardener-extension-provider-equinix-metal/pkg/apis/equinixmetal"
	"github.com/gardener/gardener-extension-provider-equinix-metal/pkg/equinixmetal"
)

func shootAccessSecretsFunc(namespace string) []*gutil.AccessSecret {
	return []*gutil.AccessSecret{
		gutil.NewShootAccessSecret(equinixmetal.CloudControllerManagerImageName, namespace),
	}
}

var controlPlaneChart = &chart.Chart{
	Name:       "seed-controlplane",
	EmbeddedFS: charts.InternalChart,
	Path:       filepath.Join(charts.InternalChartsPath, "seed-controlplane"),
	SubCharts: []*chart.Chart{
		{
			Name:   "cloud-provider-equinix-metal",
			Images: []string{equinixmetal.CloudControllerManagerImageName},
			Objects: []*chart.Object{
				{Type: &corev1.Service{}, Name: "cloud-controller-manager"},
				{Type: &appsv1.Deployment{}, Name: "cloud-controller-manager"},
				{Type: &corev1.ConfigMap{}, Name: "cloud-controller-manager-observability-config"},
			},
		},
	},
}

var controlPlaneShootChart = &chart.Chart{
	Name:       "shoot-system-components",
	EmbeddedFS: charts.InternalChart,
	Path:       filepath.Join(charts.InternalChartsPath, "shoot-system-components"),
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
			Images:  []string{equinixmetal.MetalLBControllerImageName, equinixmetal.MetalLBSpeakerImageName},
			Objects: []*chart.Object{},
		},
	},
}

var storageClassChart = &chart.Chart{
	Name:       "shoot-storageclasses",
	EmbeddedFS: charts.InternalChart,
	Path:       filepath.Join(charts.InternalChartsPath, "shoot-storageclasses"),
}

// NewValuesProvider creates a new ValuesProvider for the generic actuator.
func NewValuesProvider(mgr manager.Manager) genericactuator.ValuesProvider {
	return &valuesProvider{
		client:  mgr.GetClient(),
		decoder: serializer.NewCodecFactory(mgr.GetScheme(), serializer.EnableStrict).UniversalDecoder(),
	}
}

// valuesProvider is a ValuesProvider that provides Equinix Metal-specific values for the 2 charts applied by the generic actuator.
type valuesProvider struct {
	genericactuator.NoopValuesProvider
	client  client.Client
	decoder runtime.Decoder
}

// GetControlPlaneChartValues returns the values for the control plane chart applied by the generic actuator.
func (vp *valuesProvider) GetControlPlaneChartValues(
	_ context.Context,
	cp *extensionsv1alpha1.ControlPlane,
	cluster *extensionscontroller.Cluster,
	_ secretsmanager.Reader,
	checksums map[string]string,
	scaledDown bool,
) (
	map[string]interface{},
	error,
) {
	// Get control plane chart values
	return getControlPlaneChartValues(cp, cluster, checksums, scaledDown)
}

// GetControlPlaneShootChartValues returns the values for the control plane shoot chart applied by the generic actuator.
func (vp *valuesProvider) GetControlPlaneShootChartValues(
	_ context.Context,
	_ *extensionsv1alpha1.ControlPlane,
	cluster *extensionscontroller.Cluster,
	_ secretsmanager.Reader,
	_ map[string]string,
) (map[string]interface{}, error) {
	return getControlPlaneShootChartValues(cluster)
}

// getCredentials determines the credentials from the secret referenced in the ControlPlane resource.
func (vp *valuesProvider) getCredentials(
	ctx context.Context,
	cp *extensionsv1alpha1.ControlPlane,
) (*equinixmetal.Credentials, error) {
	secret, err := extensionscontroller.GetSecretByReference(ctx, vp.client, &cp.Spec.SecretRef)
	if err != nil {
		return nil, errors.Wrapf(err, "could not get secret by reference for controlplane '%s'", kutil.ObjectName(cp))
	}
	credentials, err := equinixmetal.ReadCredentialsSecret(secret)
	if err != nil {
		return nil, errors.Wrapf(err, "could not read credentials from secret '%s'", kutil.ObjectName(secret))
	}
	return credentials, nil
}

// getControlPlaneChartValues collects and returns the control plane chart values.
func getControlPlaneChartValues(
	cp *extensionsv1alpha1.ControlPlane,
	cluster *extensionscontroller.Cluster,
	checksums map[string]string,
	scaledDown bool,
) (
	map[string]interface{},
	error,
) {
	values := map[string]interface{}{
		"global": map[string]interface{}{
			"genericTokenKubeconfigSecretName": extensionscontroller.GenericTokenKubeconfigSecretNameFromCluster(cluster),
		},
		"cloud-provider-equinix-metal": map[string]interface{}{
			"replicas":    extensionscontroller.GetControlPlaneReplicas(cluster, scaledDown, 1),
			"clusterName": cp.Namespace,
			"podNetwork":  extensionscontroller.GetPodNetwork(cluster),
			"podAnnotations": map[string]interface{}{
				"checksum/secret-cloudprovider": checksums[v1beta1constants.SecretNameCloudProvider],
			},
			"metro": cluster.Shoot.Spec.Region,
		},
		"metallb": map[string]interface{}{},
	}

	return values, nil
}

// getControlPlaneShootChartValues collects and returns the control plane shoot chart values.
func getControlPlaneShootChartValues(
	cluster *extensionscontroller.Cluster,
) (map[string]interface{}, error) {
	return map[string]interface{}{
		"metallb": map[string]interface{}{},
	}, nil
}

func (vp *valuesProvider) decodeControlPlaneConfig(cp *extensionsv1alpha1.ControlPlane) (*api.ControlPlaneConfig, error) {
	cpConfig := &api.ControlPlaneConfig{}

	if cp.Spec.ProviderConfig == nil {
		return cpConfig, nil
	}

	if _, _, err := vp.decoder.Decode(cp.Spec.ProviderConfig.Raw, nil, cpConfig); err != nil {
		return nil, errors.Wrapf(err, "decoding '%s'", kutil.ObjectName(cp))
	}

	return cpConfig, nil
}
