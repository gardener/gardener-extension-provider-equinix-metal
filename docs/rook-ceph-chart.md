# Rook Ceph Chart

The rook-ceph chart [here](../charts/internal/shoot-system-components/charts/rook-ceph)
is the official release chart from rook, unmodified. This makes it _much_ easier
to update changes as they come along.

The general documentation is available [here](https://github.com/rook/rook/blob/master/Documentation/helm-operator.md).

To update the chart:

1. Ensure you have [helm](https://helm.sh) installed locally
1. `cd` to the root directory of this repository
1. `rm -rf charts/internal/shoot-system-components/charts/rook-ceph`
1. Download the latest, or, if you want a particular version, pass in the `--version=<version>` option:
```console
helm pull [--version=<version>]  --untar --untardir charts/internal/shoot-system-components/charts --repo https://charts.rook.io/release rook-ceph
```
1. Update [images.yaml](../charts/images.yaml), if the underlying image version has changed.
