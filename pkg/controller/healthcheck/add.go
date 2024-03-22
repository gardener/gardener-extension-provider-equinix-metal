// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package healthcheck

import (
	"context"
	"time"

	healthconfig "github.com/gardener/gardener/extensions/pkg/apis/config"
	"github.com/gardener/gardener/extensions/pkg/controller/healthcheck"
	"github.com/gardener/gardener/extensions/pkg/controller/healthcheck/general"
	"github.com/gardener/gardener/extensions/pkg/controller/healthcheck/worker"
	extensionspredicate "github.com/gardener/gardener/extensions/pkg/predicate"
	gardencorev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"github.com/gardener/gardener-extension-provider-equinix-metal/pkg/equinixmetal"
)

var (
	defaultSyncPeriod = time.Second * 30
	// DefaultAddOptions are the default DefaultAddArgs for AddToManager.
	DefaultAddOptions = healthcheck.DefaultAddArgs{
		HealthCheckConfig: healthconfig.HealthCheckConfig{
			SyncPeriod: metav1.Duration{Duration: defaultSyncPeriod},
			ShootRESTOptions: &healthconfig.RESTOptions{
				QPS:   pointer.Float32(100),
				Burst: pointer.Int(130),
			},
		},
	}
	// GardenletManagesMCM specifies whether the machine-controller-manager is managed by gardenlet.
	GardenletManagesMCM bool
)

// RegisterHealthChecks registers health checks for each extension resource
// HealthChecks are grouped by extension (e.g worker), extension.type (e.g aws) and  Health Check Type (e.g ShootControlPlaneHealthy)
func RegisterHealthChecks(ctx context.Context, mgr manager.Manager, opts healthcheck.DefaultAddArgs) error {
	if err := healthcheck.DefaultRegistration(
		ctx,
		equinixmetal.Type,
		extensionsv1alpha1.SchemeGroupVersion.WithKind(extensionsv1alpha1.ControlPlaneResource),
		func() client.ObjectList { return &extensionsv1alpha1.ControlPlaneList{} },
		func() extensionsv1alpha1.Object { return &extensionsv1alpha1.ControlPlane{} },
		mgr,
		opts,
		[]predicate.Predicate{extensionspredicate.HasPurpose(extensionsv1alpha1.Normal)},
		[]healthcheck.ConditionTypeToHealthCheck{
			{
				ConditionType: string(gardencorev1beta1.ShootControlPlaneHealthy),
				HealthCheck:   general.NewSeedDeploymentHealthChecker(equinixmetal.CloudControllerManagerName),
			},
		},
		sets.Set[gardencorev1beta1.ConditionType]{},
	); err != nil {
		return err
	}

	var (
		workerHealthChecks = []healthcheck.ConditionTypeToHealthCheck{{
			ConditionType: string(gardencorev1beta1.ShootEveryNodeReady),
			HealthCheck:   worker.NewNodesChecker(),
		}}
		workerConditionTypesToRemove = sets.New(gardencorev1beta1.ShootControlPlaneHealthy)
	)

	if !GardenletManagesMCM {
		workerHealthChecks = append(workerHealthChecks, healthcheck.ConditionTypeToHealthCheck{
			ConditionType: string(gardencorev1beta1.ShootControlPlaneHealthy),
			HealthCheck:   general.NewSeedDeploymentHealthChecker(equinixmetal.MachineControllerManagerName),
		})
		workerConditionTypesToRemove = workerConditionTypesToRemove.Delete(gardencorev1beta1.ShootControlPlaneHealthy)
	}

	return healthcheck.DefaultRegistration(
		ctx,
		equinixmetal.Type,
		extensionsv1alpha1.SchemeGroupVersion.WithKind(extensionsv1alpha1.WorkerResource),
		func() client.ObjectList { return &extensionsv1alpha1.WorkerList{} },
		func() extensionsv1alpha1.Object { return &extensionsv1alpha1.Worker{} },
		mgr,
		opts,
		nil,
		workerHealthChecks,
		workerConditionTypesToRemove,
	)
}

// AddToManager adds a controller with the default Options.
func AddToManager(ctx context.Context, mgr manager.Manager) error {
	return RegisterHealthChecks(ctx, mgr, DefaultAddOptions)
}
