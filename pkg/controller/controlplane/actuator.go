// Copyright (c) 2021 SAP SE or an SAP affiliate company. All rights reserved.
// This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
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

	"github.com/gardener/gardener-extension-provider-equinix-metal/pkg/imagevector"
	extensionscontroller "github.com/gardener/gardener/extensions/pkg/controller"
	"github.com/gardener/gardener/extensions/pkg/controller/controlplane"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/gardener/gardener/pkg/utils/chart"
	"github.com/gardener/gardener/pkg/utils/managedresources"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/runtime/inject"
)

// NewActuator creates a new Actuator that acts upon and updates the status of ControlPlane resources.
func NewActuator(
	a controlplane.Actuator,
	additionalCharts []*chart.Chart,
	chartRendererFactory extensionscontroller.ChartRendererFactory,
) controlplane.Actuator {
	return &actuator{
		Actuator:             a,
		additionalCharts:     additionalCharts,
		chartRendererFactory: chartRendererFactory,
	}
}

// actuator is a controlplane.Actuator that acts upon and updates the status of ControlPlane resources.
// This one runs the normal control plane activities, and then also applies the additional charts, if any
// chart.
type actuator struct {
	controlplane.Actuator
	client                         client.Client
	chartRendererFactory           extensionscontroller.ChartRendererFactory
	additionalChartsValuesProvider valuesProvider
	additionalCharts               []*chart.Chart
}

// InjectFunc enables injecting Kubernetes dependencies into actuator's dependencies.
func (a *actuator) InjectFunc(f inject.Func) error {
	return f(a.Actuator)
}

// InjectClient injects the given client into the valuesProvider.
func (a *actuator) InjectClient(client client.Client) error {
	a.client = client
	return nil
}

// Reconcile reconcile the core controlplane chart, and then any additional charts.
func (a *actuator) Reconcile(
	ctx context.Context,
	cp *extensionsv1alpha1.ControlPlane,
	cluster *extensionscontroller.Cluster,
) (bool, error) {
	// Call Reconcile on the composed Actuator
	requeue, err := a.Actuator.Reconcile(ctx, cp, cluster)
	if err != nil {
		return false, err
	}

	ivector := imagevector.ImageVector()
	version := cluster.Shoot.Spec.Kubernetes.Version

	// Create shoot chart renderer
	chartRenderer, err := a.chartRendererFactory.NewChartRendererForShoot(version)
	if err != nil {
		return false, fmt.Errorf("could not create chart renderer for shoot '%s': %w", cp.Namespace, err)
	}

	// Get shoot additional components charts values
	values, err := a.additionalChartsValuesProvider.GetShootAdditionalChartValues(cp, cluster)
	if err != nil {
		return false, err
	}

	// Render each shoot additional components chart and create a managed resource
	for _, chart := range a.additionalCharts {
		// Render each shoot additional components chart and create a managed resource.
		// We extract any namespace override from the values for the chart.
		chartValues, hasChartValues := values[chart.Name]
		// each additional chart must be enabled explicitly
		if !hasChartValues {
			continue
		}
		if enabledVal, ok := chartValues["enabled"]; !ok || enabledVal == nil {
			continue
		} else if enabled, ok := enabledVal.(bool); !ok {
			return false, fmt.Errorf("shoot additional components enabled for chart %s was not boolean, instead: %v", chart.Name, enabledVal)
		} else if !enabled {
			continue
		}
		var chartNamespace string
		if ns, ok := chartValues["namespace"]; !ok {
			chartNamespace = ""
		} else if chartNamespace, ok = ns.(string); !ok {
			return false, fmt.Errorf("shoot additional components namespace for chart %s was not blank and not a string, instead: %v", chart.Name, ns)
		}

		if err := managedresources.RenderChartAndCreate(ctx, cp.Namespace, chart.Name, false, a.client, chartRenderer, chart, values[chart.Name], ivector, chartNamespace, version, true, false); err != nil {
			return false, err
		}
	}
	return requeue, nil
}

// Delete reconciles the given controlplane and cluster, deleting the additional
// control plane components as needed.
// Before delegating to the composed Actuator, it ensures that all additional charts have been deleted.
// It does it in the reverse order of creation
func (a *actuator) Delete(
	ctx context.Context,
	cp *extensionsv1alpha1.ControlPlane,
	cluster *extensionscontroller.Cluster,
) error {
	// delete our additional charts, then the namespace, then the core controlplane
	// Render each shoot additional components chart and create a managed resource
	for _, chart := range a.additionalCharts {
		if err := managedresources.Delete(ctx, a.client, cp.Namespace, chart.Name, false); err != nil {
			return err
		}
	}

	// then delete our core controlplane
	return a.Actuator.Delete(ctx, cp, cluster)
}
