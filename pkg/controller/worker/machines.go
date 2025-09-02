// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package worker

import (
	"context"
	"fmt"
	"path/filepath"

	extensionscontroller "github.com/gardener/gardener/extensions/pkg/controller"
	"github.com/gardener/gardener/extensions/pkg/controller/worker"
	genericworkeractuator "github.com/gardener/gardener/extensions/pkg/controller/worker/genericactuator"
	v1beta1constants "github.com/gardener/gardener/pkg/apis/core/v1beta1/constants"
	gardencorev1beta1helper "github.com/gardener/gardener/pkg/apis/core/v1beta1/helper"
	extensionsv1alpha1helper "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1/helper"
	"github.com/gardener/gardener/pkg/client/kubernetes"
	machinev1alpha1 "github.com/gardener/machine-controller-manager/pkg/apis/machine/v1alpha1"
	"k8s.io/utils/ptr"

	"github.com/gardener/gardener-extension-provider-equinix-metal/charts"
	api "github.com/gardener/gardener-extension-provider-equinix-metal/pkg/apis/equinixmetal"
	"github.com/gardener/gardener-extension-provider-equinix-metal/pkg/equinixmetal"
)

// DeployMachineClasses generates and creates the Equinix Metal specific machine classes.
func (w *workerDelegate) DeployMachineClasses(ctx context.Context) error {
	if w.machineClasses == nil {
		if err := w.generateMachineConfig(ctx); err != nil {
			return err
		}
	}

	return w.seedChartApplier.ApplyFromEmbeddedFS(ctx, charts.InternalChart, filepath.Join(charts.InternalChartsPath, "machineclass"), w.worker.Namespace, "machineclass", kubernetes.Values(map[string]interface{}{"machineClasses": w.machineClasses}))
}

// GenerateMachineDeployments generates the configuration for the desired machine deployments.
func (w *workerDelegate) GenerateMachineDeployments(ctx context.Context) (worker.MachineDeployments, error) {
	if w.machineDeployments == nil {
		if err := w.generateMachineConfig(ctx); err != nil {
			return nil, err
		}
	}
	return w.machineDeployments, nil
}

func (w *workerDelegate) generateMachineConfig(ctx context.Context) error {
	var (
		machineDeployments = worker.MachineDeployments{}
		machineClasses     []map[string]interface{}
		machineImages      []api.MachineImage
	)

	infrastructureStatus := &api.InfrastructureStatus{}
	if _, _, err := w.decoder.Decode(w.worker.Spec.InfrastructureProviderStatus.Raw, nil, infrastructureStatus); err != nil {
		return err
	}

	secret, err := extensionscontroller.GetSecretByReference(ctx, w.client, &w.worker.Spec.SecretRef)
	if err != nil {
		return err
	}
	credentials, err := equinixmetal.ReadCredentialsSecret(secret)
	if err != nil {
		return err
	}

	for _, pool := range w.worker.Spec.Pools {
		workerConfig := &api.WorkerConfig{}
		if pool.ProviderConfig != nil && pool.ProviderConfig.Raw != nil {
			if _, _, err := w.decoder.Decode(pool.ProviderConfig.Raw, nil, workerConfig); err != nil {
				return fmt.Errorf("could not decode provider config: %+v", err)
			}
		}

		workerPoolHash, err := worker.WorkerPoolHash(pool, w.cluster, []string{}, []string{})
		if err != nil {
			return err
		}

		machineImage, err := w.findMachineImage(pool.MachineImage.Name, pool.MachineImage.Version)
		if err != nil {
			return err
		}
		machineImages = appendMachineImage(machineImages, api.MachineImage{
			Name:          pool.MachineImage.Name,
			Version:       pool.MachineImage.Version,
			ID:            machineImage.ID,
			IPXEScriptURL: machineImage.IPXEScriptURL,
		})

		userData, err := worker.FetchUserData(ctx, w.client, w.worker.Namespace, pool)
		if err != nil {
			return err
		}

		machineClassSpec := map[string]interface{}{
			"OS":            machineImage.ID,
			"ipxeScriptUrl": machineImage.IPXEScriptURL,
			"projectID":     string(credentials.ProjectID),
			"billingCycle":  "hourly",
			"machineType":   pool.MachineType,
			"metro":         w.worker.Spec.Region,
			"sshKeys":       []string{infrastructureStatus.SSHKeyID},
			"tags": []string{
				fmt.Sprintf("kubernetes.io/cluster/%s", w.worker.Namespace),
				"kubernetes.io/role/node",
			},
			"secret": map[string]interface{}{
				"cloudConfig": string(userData),
			},
			"credentialsSecretRef": map[string]interface{}{
				"name":      w.worker.Spec.SecretRef.Name,
				"namespace": w.worker.Spec.SecretRef.Namespace,
			},
		}

		if len(pool.Zones) > 0 {
			machineClassSpec["facilities"] = pool.Zones
		}

		if len(workerConfig.ReservationIDs) > 0 {
			machineClassSpec["reservationIDs"] = workerConfig.ReservationIDs
		}

		if workerConfig.ReservedDevicesOnly != nil {
			machineClassSpec["reservedDevicesOnly"] = *workerConfig.ReservedDevicesOnly
		}

		var (
			deploymentName = fmt.Sprintf("%s-%s", w.worker.Namespace, pool.Name)
			className      = fmt.Sprintf("%s-%s", deploymentName, workerPoolHash)
		)

		updateConfiguration := machinev1alpha1.UpdateConfiguration{
			MaxUnavailable: ptr.To(pool.MaxUnavailable),
			MaxSurge:       ptr.To(pool.MaxSurge),
		}

		machineDeploymentStrategy := machinev1alpha1.MachineDeploymentStrategy{
			Type: machinev1alpha1.RollingUpdateMachineDeploymentStrategyType,
			RollingUpdate: &machinev1alpha1.RollingUpdateMachineDeployment{
				UpdateConfiguration: updateConfiguration,
			},
		}

		if gardencorev1beta1helper.IsUpdateStrategyInPlace(pool.UpdateStrategy) {
			machineDeploymentStrategy = machinev1alpha1.MachineDeploymentStrategy{
				Type: machinev1alpha1.InPlaceUpdateMachineDeploymentStrategyType,
				InPlaceUpdate: &machinev1alpha1.InPlaceUpdateMachineDeployment{
					UpdateConfiguration: updateConfiguration,
					OrchestrationType:   machinev1alpha1.OrchestrationTypeAuto,
				},
			}

			if gardencorev1beta1helper.IsUpdateStrategyManualInPlace(pool.UpdateStrategy) {
				machineDeploymentStrategy.InPlaceUpdate.OrchestrationType = machinev1alpha1.OrchestrationTypeManual
			}
		}

		machineDeployments = append(machineDeployments, worker.MachineDeployment{
			Name:                         deploymentName,
			ClassName:                    className,
			SecretName:                   className,
			Minimum:                      pool.Minimum,
			Maximum:                      pool.Maximum,
			Strategy:                     machineDeploymentStrategy,
			Priority:                     pool.Priority,
			Labels:                       pool.Labels,
			Annotations:                  pool.Annotations,
			Taints:                       pool.Taints,
			MachineConfiguration:         genericworkeractuator.ReadMachineConfiguration(pool),
			ClusterAutoscalerAnnotations: extensionsv1alpha1helper.GetMachineDeploymentClusterAutoscalerAnnotations(pool.ClusterAutoscaler),
		})

		machineClassSpec["name"] = className
		machineClassSpec["labels"] = map[string]string{
			v1beta1constants.GardenerPurpose: v1beta1constants.GardenPurposeMachineClass,
		}

		machineClasses = append(machineClasses, machineClassSpec)
	}

	w.machineDeployments = machineDeployments
	w.machineClasses = machineClasses
	w.machineImages = machineImages

	return nil
}
