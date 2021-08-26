# Using the Equinix Metal provider extension with Gardener as operator

The [`core.gardener.cloud/v1alpha1.CloudProfile` resource](https://github.com/gardener/gardener/blob/master/example/30-cloudprofile.yaml) declares a `providerConfig` field that is meant to contain provider-specific configuration.

In this document we are describing how this configuration looks like for Equinix Metal and provide an example `CloudProfile` manifest with minimal configuration that you can use to allow creating Equinix Metal shoot clusters.

## Example `CloudProfile` manifest

Please find below an example `CloudProfile` manifest:

```yaml
apiVersion: core.gardener.cloud/v1beta1
kind: CloudProfile
metadata:
  name: equinix-metal
spec:
  type: equinixmetal
  kubernetes:
    versions:
    - version: 1.20.2
    - version: 1.19.7
    - version: 1.18.15
      #expirationDate: "2020-04-05T01:02:03Z"
  machineImages:
  - name: flatcar
    versions:
    - version: 0.0.0-stable
  machineTypes:
  - name: t1.small
    cpu: "4"
    gpu: "0"
    memory: 8Gi
    usable: true
  regions: # List of offered metros
  - name: ny
    zones: # List of offered facilities within the respective metro
    - name: ewr1
    - name: ny5
    - name: ny7
  providerConfig:
    apiVersion: equinixmetal.provider.extensions.gardener.cloud/v1alpha1
    kind: CloudProfileConfig
    machineImages:
    - name: flatcar
      versions:
      - version: 0.0.0-stable
        id: flatcar_stable
```

## `CloudProfileConfig`

The cloud profile configuration contains information about the real machine image IDs in the Equinix Metal environment (IDs).
You have to map every version that you specify in `.spec.machineImages[].versions` here such that the Equinix Metal extension knows the ID for every version you want to offer.

An example `CloudProfileConfig` for the Equinix Metal extension looks as follows:

```yaml
apiVersion: equinixmetal.provider.extensions.gardener.cloud/v1alpha1
kind: CloudProfileConfig
machineImages:
- name: flatcar
  versions:
  - version: 0.0.0-stable
    id: flatcar_stable
```

> NOTE: `CloudProfileConfig` is not a Custom Resource, so you cannot create it directly.
