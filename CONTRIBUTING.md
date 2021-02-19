# Contributing

As a starting point, please refer to the [Gardener contributor guide](https://github.com/gardener/documentation/blob/master/CONTRIBUTING.md).

## Equinix Metal Extension Provider

The rest of this document describes how to contribute to, and test, this Equinix Metal (formerly Packet) extension provider for Gardener.

The guide demonstrates how to make changes, test them and publish them using the
[Packet provider extension](https://github.com/gardener/gardener-extension-provider-packet).

Gardener uses
[an extensible architecture](https://github.com/gardener/gardener/blob/master/docs/proposals/01-extensibility.md) 
which abstracts away things like cloud provider
implementations, OS-specific configuration for nodes and DNS provider logic.

Extensions are k8s controllers. They are *packaged* as Helm charts and are *registered* with the
Gardener API server by applying a `ControllerRegistration` custom k8s resource to the Gardener API
server.
In addition to being packaged as Helm charts, extensions often **deploy** Helm charts as well. For
example, the Packet provider extension deploys components such as the Packet CCM and MetalLB in
order to provide necessary services to Packet seed and shoot clusters.

## Requirements

- kubectl access to a running "garden" cluster

## Development workflow

**Important!** Gardener uses a special k8s API server which doesn't have any nodes or pods and
which deals with Gardener resources only. In this guide we refer to this API server as the
"Gardener API server". The kubeconfig file for accessing this API server is generated when creating
a garden cluster into `export/kube-apiserver/kubeconfig` in the "landscape" directory. The
"regular" API server that is used by the underlying k8s cluster is a *different* API server and we
refer to it as the "base API server".

### Deploying an extension to a "garden" cluster

>NOTE: Some operations in the development workflow don't support Go modules. It is recommended to
>clone the extension's source directory into your GOPATH.

Ensure the parent directory under the GOPATH exist, then clone the extension's source repository:

```console
mkdir -p $GOPATH/src/github.com/gardener && cd $_
git clone git@github.com:gardener/gardener-extension-provider-packet.git
cd gardener-extension-provider-packet
```

Set your `KUBECONFIG` env var to point at the kubeconfig file belonging to the **Gardener** API
server (not the base API server). For a Gardener cluster deployed using
[garden-setup](https://github.com/gardener/garden-setup), the command should be the following:

```
export KUBECONFIG=$(pwd)/export/kube-apiserver/kubeconfig
```

Verify connectivity to the Gardener API server:

```
kubectl get seeds
```

Sample output:

```
NAME   STATUS   PROVIDER   REGION         AGE    VERSION   K8S VERSION
aws    Ready    aws        eu-central-1   104m   v1.14.0   v1.18.12
```

>NOTE: The `controller-registration.yaml` file in the `master` branch contains a reference to an
>**unreleased** container image. Either overwrite the `tag` field in the file or check out a Git
>tag before registering the extension.

Register the extension:

```
kubectl apply -f ./example/controller-registration.yaml
```

From this point the extension should be deployed on-demand, e.g. when a shoot cluster gets created.

### Making changes to an extension

Install the development requirements:

```
make install-requirements
```

Make code changes, then run the following:

```
# Run the tests
make test

# Re-generate any auto-generated files
make generate

# Verify auto-generated files are up to date
make check-generate

# Build a Docker image
make docker-images
```

You can now tag and push the resulting image to a Docker registry:

```
docker tag 9fa4c197cf04 quay.io/jlieb/gardener-extension-packet:testing
docker push quay.io/jlieb/gardener-extension-packet:testing
```

To use the new image, register or re-register the extension with the Gardener API server. Note that
when re-pushing an existing Docker tag, you will likely have to temporarily set the extension's
`imagePullPolicy` field to `Always` and delete the extension's pod to force the kubelet to re-pull
the image.

### Debugging

To check the logs of the extension controller, run the following command against the **base** API
server:

```
kubectl -n extension-provider-packet-xxxxx logs gardener-extension-provider-packet-xxxxx
```

### Miscellaneous and "gotchas"

The Helm chart that is used by Gardener to deploy the extension controller is packaged as a
**base64-encoded gzip archive** inline inside `example/controller-registration.yaml` under the
`chart` field. To view its contents, run the following command:

```
cat example/controller-registration.yaml | grep chart | awk {'print $2'} | base64 -d | tar zxvf -
```

Some generated files are **gitignored**. To avoid building images using outdated generated data, be
sure to **always** run `make generate` before running `make docker-images` (even when there is no
visible Git diff).
