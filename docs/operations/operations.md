# Using the Equinix Metal provider extension with Gardener as operator

The [`core.gardener.cloud/v1beta1.CloudProfile` resource](https://github.com/gardener/gardener/blob/master/example/30-cloudprofile.yaml) declares a `providerConfig` field that is meant to contain provider-specific configuration.

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
    - version: 1.27.2
    - version: 1.26.7
    - version: 1.25.10
      #expirationDate: "2023-03-15T23:59:59Z"
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
      - version: 3510.2.2
        ipxeScriptUrl: https://stable.release.flatcar-linux.net/amd64-usr/3510.2.2/flatcar_production_packet.ipxe
```

## `CloudProfileConfig`

The cloud profile configuration contains information about the real machine image IDs in the Equinix Metal environment (IDs).
You have to map every version that you specify in `.spec.machineImages[].versions` here such that the Equinix Metal extension knows the ID for every version you want to offer.

Equinix Metal supports two different options to specify the image:
1. Supported Operating System: Images that are provided by Equinix Metal. They are referenced by their ID (`slug`). See (Operating Systems Reference)[https://deploy.equinix.com/developers/docs/metal/operating-systems/supported/#operating-systems-reference] for all supported operating system and their ids.
2. Custom iPXE Boot: Equinix Metal supports passing custom iPXE scripts during provisioning, which allows you to install a custom operating system manually. This is useful if you want to have a custom image or want to pin to a specific version. See [Custom iPXE Boot](https://deploy.equinix.com/developers/docs/metal/operating-systems/custom-ipxe/#provisioning-with-custom-ipxe) for details.

An example `CloudProfileConfig` for the Equinix Metal extension looks as follows:

```yaml
apiVersion: equinixmetal.provider.extensions.gardener.cloud/v1alpha1
kind: CloudProfileConfig
machineImages:
- name: flatcar
  versions:
  - version: 0.0.0-stable
    id: flatcar_stable
  - version: 3510.2.2
    ipxeScriptUrl: https://stable.release.flatcar-linux.net/amd64-usr/3510.2.2/flatcar_production_packet.ipxe
```

> NOTE: `CloudProfileConfig` is not a Custom Resource, so you cannot create it directly.
