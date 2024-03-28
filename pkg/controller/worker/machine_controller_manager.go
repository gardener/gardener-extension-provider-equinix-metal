// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package worker

import (
	"context"
	"fmt"
	"path/filepath"

	extensionscontroller "github.com/gardener/gardener/extensions/pkg/controller"
	"github.com/gardener/gardener/pkg/utils/chart"
	kutil "github.com/gardener/gardener/pkg/utils/kubernetes"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"

	"github.com/gardener/gardener-extension-provider-equinix-metal/charts"
	"github.com/gardener/gardener-extension-provider-equinix-metal/pkg/equinixmetal"
)

var (
	mcmChart = &chart.Chart{
		Name:       equinixmetal.MachineControllerManagerName,
		EmbeddedFS: &charts.InternalChart,
		Path:       filepath.Join(charts.InternalChartsPath, equinixmetal.MachineControllerManagerName, "seed"),
		Images:     []string{equinixmetal.MachineControllerManagerImageName, equinixmetal.MachineControllerManagerEquinixMetalImageName},
		Objects: []*chart.Object{
			{Type: &appsv1.Deployment{}, Name: equinixmetal.MachineControllerManagerName},
			{Type: &corev1.Service{}, Name: equinixmetal.MachineControllerManagerName},
			{Type: &corev1.ServiceAccount{}, Name: equinixmetal.MachineControllerManagerName},
			{Type: &corev1.Secret{}, Name: equinixmetal.MachineControllerManagerName},
			{Type: extensionscontroller.GetVerticalPodAutoscalerObject(), Name: equinixmetal.MachineControllerManagerVpaName},
			{Type: &corev1.ConfigMap{}, Name: equinixmetal.MachineControllerManagerMonitoringConfigName},
		},
	}

	mcmShootChart = &chart.Chart{
		Name:       equinixmetal.MachineControllerManagerName,
		EmbeddedFS: &charts.InternalChart,
		Path:       filepath.Join(charts.InternalChartsPath, equinixmetal.MachineControllerManagerName, "shoot"),
		Objects: []*chart.Object{
			{Type: &rbacv1.ClusterRole{}, Name: fmt.Sprintf("extensions.gardener.cloud:%s:%s", equinixmetal.Name, equinixmetal.MachineControllerManagerName)},
			{Type: &rbacv1.ClusterRoleBinding{}, Name: fmt.Sprintf("extensions.gardener.cloud:%s:%s", equinixmetal.Name, equinixmetal.MachineControllerManagerName)},
		},
	}
)

func (w *workerDelegate) GetMachineControllerManagerChartValues(ctx context.Context) (map[string]interface{}, error) {
	namespace := &corev1.Namespace{}
	if err := w.client.Get(ctx, kutil.Key(w.worker.Namespace), namespace); err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"providerName": equinixmetal.Name,
		"namespace": map[string]interface{}{
			"uid": namespace.UID,
		},
	}, nil
}

func (w *workerDelegate) GetMachineControllerManagerShootChartValues(ctx context.Context) (map[string]interface{}, error) {
	return map[string]interface{}{
		"providerName": equinixmetal.Name,
	}, nil
}
