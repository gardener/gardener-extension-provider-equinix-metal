# Using the Equinix Metal provider extension with Gardener as end-user

The [`core.gardener.cloud/v1beta1.Shoot` resource](https://github.com/gardener/gardener/blob/master/example/90-shoot.yaml) declares a few fields that are meant to contain provider-specific configuration.

In this document we are describing how this configuration looks like for Equinix Metal and provide an example `Shoot` manifest with minimal configuration that you can use to create an Equinix Metal cluster (modulo the landscape-specific information like cloud profile names, secret binding names, etc.).

## Provider secret data

Every shoot cluster references a `SecretBinding` which itself references a `Secret`, and this `Secret` contains the provider credentials of your Equinix Metal project.
This `Secret` must look as follows:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: my-secret
  namespace: garden-dev
type: Opaque
data:
  apiToken: base64(api-token)
  projectID: base64(project-id)
```

Please look up https://metal.equinix.com/developers/api/ as well.

With `Secret` created, create a `SecretBinding` resource referencing it. It may look like this:

```yaml
apiVersion: core.gardener.cloud/v1beta1
kind: SecretBinding
metadata:
  name: my-secret
  namespace: garden-dev
secretRef:
  name: my-secret
quotas: []
```

## `InfrastructureConfig`

Currently, there is no infrastructure configuration possible for the Equinix Metal environment.

An example `InfrastructureConfig` for the Equinix Metal extension looks as follows:

```yaml
apiVersion: equinixmetal.provider.extensions.gardener.cloud/v1alpha1
kind: InfrastructureConfig
```

The Equinix Metal extension will only create a key pair.

## `ControlPlaneConfig`

The control plane configuration mainly contains values for the Equinix Metal-specific control plane components.
Today, the Equinix Metal extension deploys the `cloud-controller-manager` and the CSI controllers, however, it doesn't offer any configuration options at the moment.

An example `ControlPlaneConfig` for the Equinix Metal extension looks as follows:

```yaml
apiVersion: equinixmetal.provider.extensions.gardener.cloud/v1alpha1
kind: ControlPlaneConfig
```

## `WorkerConfig`

The Equinix Metal extension supports specifying IDs for reserved devices that should be used for the machines of a specific worker pool.

An example `WorkerConfig` for the Equinix Metal extension looks as follows:

```yaml
apiVersion: equinixmetal.provider.extensions.gardener.cloud/v1alpha1
kind: WorkerConfig
reservationIDs:
- my-reserved-device-1
- my-reserved-device-2
reservedDevicesOnly: false
```

The `.reservationIDs[]` list contains the list of IDs of the reserved devices.
The `.reservedDevicesOnly` field indicates whether only reserved devices from the provided list of reservation IDs should be used when new machines are created.
It always will attempt to create a device from one of the reservation IDs.
If none is available, the behaviour depends on the setting:

* `true`: return an error
* `false`: request a regular on-demand device

The default value is `false`.

## Example `Shoot` manifest

Please find below an example `Shoot` manifest:

```yaml
apiVersion: core.gardener.cloud/v1beta1
kind: Shoot
metadata:
  name: my-shoot
  namespace: garden-dev
spec:
  cloudProfileName: equinix-metal
  region: ny # Corresponds to a metro
  secretBindingName: my-secret
  provider:
    type: equinixmetal
    infrastructureConfig:
      apiVersion: equinixmetal.provider.extensions.gardener.cloud/v1alpha1
      kind: InfrastructureConfig
    controlPlaneConfig:
      apiVersion: equinixmetal.provider.extensions.gardener.cloud/v1alpha1
      kind: ControlPlaneConfig
    workers:
    - name: worker-pool1
      machine:
        type: t1.small
      minimum: 2
      maximum: 2
      volume:
        size: 50Gi
        type: storage_1
      zones: # Optional list of facilities, all of which MUST be in the metro; if not provided, then random facilities within the metro will be chosen for each machine.
      - ewr1
      - ny5
    - name: reserved-pool
      machine:
        type: t1.small
      minimum: 1
      maximum: 2
      providerConfig:
        apiVersion: equinixmetal.provider.extensions.gardener.cloud/v1alpha1
        kind: WorkerConfig
        reservationIDs:
        - reserved-device1
        - reserved-device2
        reservedDevicesOnly: true
      volume:
        size: 50Gi
        type: storage_1
  networking:
    type: calico
  kubernetes:
    version: 1.27.2
  maintenance:
    autoUpdate:
      kubernetesVersion: true
      machineImageVersion: true
  addons:
    kubernetesDashboard:
      enabled: true
    nginxIngress:
      enabled: true
```

⚠️ Note that if you specify multiple facilities in the `.spec.provider.workers[].zones[]` list then new machines are randomly created in one of the provided facilities.
Particularly, it is not ensured that all facilities are used or that all machines are equally or unequally distributed.

## Kubernetes Versions per Worker Pool

This extension supports `gardener/gardener`'s `WorkerPoolKubernetesVersion` feature gate, i.e., having [worker pools with overridden Kubernetes versions](https://github.com/gardener/gardener/blob/8a9c88866ec5fce59b5acf57d4227eeeb73669d7/example/90-shoot.yaml#L69-L70) since `gardener-extension-provider-equinix-metal@v2.2`.

## Shoot CA Certificate and `ServiceAccount` Signing Key Rotation

This extension supports `gardener/gardener`'s `ShootCARotation` feature gate since `gardener-extension-provider-equinix-metal@v2.3` and `ShootSARotation` feature gate since `gardener-extension-provider-equinix-metal@v2.4`.
