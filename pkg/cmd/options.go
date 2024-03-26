// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	controllercmd "github.com/gardener/gardener/extensions/pkg/controller/cmd"
	extensionscontrolplanecontroller "github.com/gardener/gardener/extensions/pkg/controller/controlplane"
	extensionshealthcheckcontroller "github.com/gardener/gardener/extensions/pkg/controller/healthcheck"
	extensionsheartbeatcontroller "github.com/gardener/gardener/extensions/pkg/controller/heartbeat"
	extensionsinfrastructurecontroller "github.com/gardener/gardener/extensions/pkg/controller/infrastructure"
	extensionsworkercontroller "github.com/gardener/gardener/extensions/pkg/controller/worker"
	webhookcmd "github.com/gardener/gardener/extensions/pkg/webhook/cmd"
	extensioncontrolplanewebhook "github.com/gardener/gardener/extensions/pkg/webhook/controlplane"
	extensionshootwebhook "github.com/gardener/gardener/extensions/pkg/webhook/shoot"

	controlplanecontroller "github.com/gardener/gardener-extension-provider-equinix-metal/pkg/controller/controlplane"
	healthcheckcontroller "github.com/gardener/gardener-extension-provider-equinix-metal/pkg/controller/healthcheck"
	infrastructurecontroller "github.com/gardener/gardener-extension-provider-equinix-metal/pkg/controller/infrastructure"
	workercontroller "github.com/gardener/gardener-extension-provider-equinix-metal/pkg/controller/worker"
	controlplanewebhook "github.com/gardener/gardener-extension-provider-equinix-metal/pkg/webhook/controlplane"
	controlplaneexposurewebhook "github.com/gardener/gardener-extension-provider-equinix-metal/pkg/webhook/controlplaneexposure"
	shootwebhook "github.com/gardener/gardener-extension-provider-equinix-metal/pkg/webhook/shoot"
)

// ControllerSwitchOptions are the controllercmd.SwitchOptions for the provider controllers.
func ControllerSwitchOptions() *controllercmd.SwitchOptions {
	return controllercmd.NewSwitchOptions(
		controllercmd.Switch(extensionsinfrastructurecontroller.ControllerName, infrastructurecontroller.AddToManager),
		controllercmd.Switch(extensionscontrolplanecontroller.ControllerName, controlplanecontroller.AddToManager),
		controllercmd.Switch(extensionsworkercontroller.ControllerName, workercontroller.AddToManager),
		controllercmd.Switch(extensionshealthcheckcontroller.ControllerName, healthcheckcontroller.AddToManager),
		controllercmd.Switch(extensionsheartbeatcontroller.ControllerName, extensionsheartbeatcontroller.AddToManager),
	)
}

// WebhookSwitchOptions are the webhookcmd.SwitchOptions for the provider webhooks.
func WebhookSwitchOptions() *webhookcmd.SwitchOptions {
	return webhookcmd.NewSwitchOptions(
		webhookcmd.Switch(extensioncontrolplanewebhook.WebhookName, controlplanewebhook.AddToManager),
		webhookcmd.Switch(extensioncontrolplanewebhook.ExposureWebhookName, controlplaneexposurewebhook.AddToManager),
		webhookcmd.Switch(extensionshootwebhook.WebhookName, shootwebhook.AddToManager),
	)
}
