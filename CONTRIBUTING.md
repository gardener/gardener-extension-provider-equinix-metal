# Contributing

As a starting point, please refer to the [Gardener contributor guide](https://github.com/gardener/documentation/blob/master/CONTRIBUTING.md).

## Equinix Metal Extension Provider

The rest of this document describes how to contribute to, and test, this Equinix Metal (formerly Packet) extension provider for Gardener.

The guide demonstrates how to make changes, test them and publish them using the
[Equinix Metal provider extension](https://github.com/gardener/gardener-extension-provider-equinix-metal).

Gardener uses
[an extensible architecture](https://github.com/gardener/gardener/blob/master/docs/proposals/01-extensibility.md)
which abstracts away things like cloud provider
implementations, OS-specific configuration for nodes and DNS provider logic.

Extensions are k8s controllers. They are *packaged* as Helm charts and are *registered* with the
Gardener API server by applying a `ControllerRegistration` custom k8s resource to the Gardener API
server.
In addition to being packaged as Helm charts, extensions often **deploy** Helm charts as well. For
example, the Equinix Metal provider extension deploys components such as the Equinix Metal CCM and MetalLB in
order to provide necessary services to Equinix Metal seed and shoot clusters.

## Requirements

- a running "garden" cluster with a kubeconfig for access oti

## Background

**Important!** Gardener's base control cluster, or "garden" cluster, really has
two **distinct** kinds of API servers, both of which are accessed via `kubectl` using
a `kubeconfig` file

* The "soil cluster", the pre-existing normal Kubernetes cluster on which Gardener is installed, turning it into the "soil cluster"
* The "garden cluster", a special k8s API server which doesn't have any nodes or pods and which deals with Gardener resources only, and runs on top of the "soil cluster"

The `kubeconfig` for the "soil cluster" is wherever you set it when creating the initial Kubernetes
cluster, for example during [garden-setup](https://github.com/gardener/garden-setup). If you used
the standard `garden-setup` flow, then the `kubeconfig` is likely at `$GOPATH/src/github.com/gardener/sow/landscape/kubeconfig`.

The `kubeconfig` for the "garden cluster", when using [garden-setup](https://github.com/gardener/garden-setup)
is at `$GOPATH/src/github.com/gardener/sow/landscape/export/kube-apiserver/kubeconfig`.

## Development workflow

Your development workflow in general is as follows. Relevant steps will be described in detail below.

### Setup

The setup steps are necessary just so that your extension can have what to work with against a garden cluster.
In general, you will do these once, although if you are changing how the extension works with some components,
e.g. the `CloudProfile` or `Shoot`, you may modify and redeploy them multiple times.

1. Deploy your base cluster
1. Convert your base cluster into a soil cluster, thus creating a garden cluster
1. Deploy a seed to your garden cluster
1. Deploy a project to your garden cluster
1. Get an Equinix Metal API key and project ID, and save them to the secret file, which also contains the secret binding
1. Deploy the secret and secret binding to the soil cluster
1. Configure and deploy a cloud profile to the soil cluster
1. Configure and deploy a shoot to the soil cluster

#### Base Cluster and Garden Cluster

Deploying your base cluster is beyond the scope of these documents. Deploy it any way that works for you.
We recommend following the instructions at [garden-setup](https://github.com/gardener/garden-setup).
The documentation there describes, as well, how to deploy a seed.

#### Ensure the Seed is Deployed

You need at least one seed deployed. For simplicity, you should deploy the seed right into the garden cluster.

1. Connect to the "gardener API server", e.g. `KUBECONFIG=./export/kube-apiserver/kubeconfig`
1. `kubectl get seed` - this should return at least one functional seed

For example:

```sh
$ KUBECONFIG=./export/kube-apiserver/kubeconfig  kubectl get seed
NAME   STATUS   PROVIDER   REGION        AGE   VERSION   K8S VERSION
gcp    Ready    gcp        us-central1   60d   v1.17.1   v1.19.9-gke.1400
```

Note that the above garden cluster is running in GCP, and so the seed is using the `PROVIDER` named `gcp`. That is fine.
The Equinix Metal extension will be invoked when deploying a shoot to Equinix Metal, and can do so from a seed
in GCP.

#### Deploy a Project

A `Project` groups together shoots and infrastructure secrets in a namespace.
A sample `Project` is available at [example/23-project.yaml](./example/23-project.yaml). Copy it over to a temporary workspace,
modify it as needed, and then apply it.

```sh
kubectl apply -f 23-project.yaml
```

Unless you actually will be using the Gardener UI, most of the RBAC entries in the file do not matter for development.
The only really important elements are:

* `name`: pick a unique one for the `Project`
* `namespace`: you will need to be consistent in using the same namespace for multiple elements

#### Secret

You need two pieces of information for Gardener to do its job against the Equinix Metal API:

* project ID - the UUID of the project in which it should manage devices
* API key - your unique API key that has the rights to create/delete devices in that project

Both of these should be placed in a Kubernetes `Secret`, as well as the `SecretBinding` that enables the extension to use them.

A sample `Secret` is available at [example/25-secret.yaml](./example/25-secret.yaml). Copy it over to a temporary workspace,
modify it as needed, and then apply it.

**Important:** The `Secret` and `SecretBinding` must be in the same `namespace` as the `Project`.

```sh
kubectl apply -f 25-secret.yaml
```

#### Cloud Profile

The `CloudProfile` is a resource that contains the list of acceptable machine types and OS images. Each
`Shoot`, when deployed, uses a specific `CloudProfile` and picks elements from it.

A sample `CloudProfile` is available at [example/26-cloudprofile.yaml](./example/26-cloudprofile.yaml)/ Copy it over to a
temporary workspace, modify as needed, and then apply it.

```sh
kubectl apply -f 26-cloudprofile.yaml
```

#### Shoot

With all of the above, you are ready to create a `Shoot`. The `Shoot` is the instructions for actually deploying
a Kubernetes cluster, managed by Gardener, on machines deployed by Gardener. The interaction with Equinix Metal
occurs via this extension provider.

A sample `Shoot` is available at [example/90-shoot.yaml](./example/90-shoot.yaml). Copy it over to a temporary workspace,
modify as needed, and then apply it.

```sh
kubectl apply -f 90-shoot.yaml
````

The `Shoot` resource is long and complex, and beyond the scope of this document. The important parts to note are:

* `namespace` - must match the namespace for the `Project` and `Secret`/`SecretBinding`
* `secretBindingName` - must match the name of the `SecretBinding`
* `cloudProfileName` - must match the name of the `CloudProfile`
* `region` - must be one of the regions in the `CloudProfile`
* `infrastructureConfig` - must conform to the infrastructure config, see [here](./example/30-infrastructure.yaml#L48-L49)
* `controlPlaneConfig` - must conform to the control plane config, see [here](./example/30-controlplane.yaml#L58-L59)

### Extension

With the setup in place, you can run your local extension against the soil cluster.

The extension actually is made up of two components, although they start as one. Both use the provided `KUBECONFIG` to connect to the
API server of the "base cluster", but their behaviours diverge from there.

* A kubernetes controller, which registers and listens for events; the entire connection is initiated as a client.
* A kubernetes [mutating webhook](https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/), which registers an http endpoint to which the `kube-apiserver` sends events. Each event requires a new connection initiated by the `kube-apiserver`

If your locally running extension does not have an IP address accessible to the `kube-apiserver`, you need some way to
open a tunnel for it. This is the job of `hook-me.sh`. `hook-me.sh` starts an [inlets tunnel](http://inlets.dev) to
connect to your IP.

Steps:

1. Get the namespace in which your extensions are running
1. Determine the port on which the hook will be listening
1. Determine the port on which inlets is listening
1. Determine the port on which the extension service `gardener-extension-provider-equinix-metal` is listening
1. Run the extension locally: `make start <options>`
1. Run the webhook tunnel: `./hack/hook-me.sh <provider> <namespace> <webhook port> <inlets port>`, where `<namespace>`, `<port>` and `<inlets port>` are as determined above, and `<provider>` is `equinix-metal`

#### Namespace

`<namespace>` is the namespace that the extension is running in on the "soil cluster". _After_ you deployed the shoot,
connect to the "soil cluster" in which the seed is running, and do `kubectl get ns`; the target namespace should have
your provider name in it.

For example:

```sh
$ kubectl get ns
NAME                                     STATUS   AGE
cert-manager                             Active   61d
default                                  Active   64d
extension-dns-external-j4psd             Active   6d21h
extension-networking-calico-xjshs        Active   6d21h
extension-os-ubuntu-nrhxn                Active   6d21h
extension-provider-gcp-bcdwd             Active   6d21h
extension-provider-equinix-metal-9r7xh   Active   6d21h
garden                                   Active   61d
kube-node-lease                          Active   64d
kube-public                              Active   64d
kube-system                              Active   64d
shoot--em--em-test                       Active   6d21h
```

In this example, it is `extension-provider-equinix-metal-9r7xh`.

#### Webhook Port

`<webhook port>` is the port on which the provider extension webhook component is listening. By default, it is `8443`,
but you can set it to anything you want.


it is set in the `Makefile` as `WEBHOOK_CONFIG_PORT`.

#### Inlets Port

`<inlets port>` is the port on which inlets is listening. The garden kube-apiserver uses the service `gardener-extension-provider-equinix-metal`
to connect to the webhook. The `targetPort` of that service must point either to:

* the in-cluster pod where the extension is running
* the in-cluster pod where inlets is listening, normally `inlets-server`

Because the `gardener-extension-provider-equinix-metal` service is deployed and syncs every 5 minutes, changing the `targetPort`
on this service to match the port on which `inlets-server` is listening will only last a few minutes.

Thus, you want the `inlets-server` listening port to match up to the `targetPort` of `gardener-extension-provider-equinix-metal`

The `gardener-extension-provider-equinix-metal` port is set in [values.yaml](./charts/gardener-extension-provider-equinix-metal/values.yaml#L45),
by default 10250, and the inlets port is set as the last argument to `hook-me.sh`.

#### Extension Service Port

The shoot kube-apiserver will be pointed to the `gardener-extension-provider-equinix-metal` Service, which, in turn, as described above, points
to the inlets pod. Thus, the apiserver must be told which port it already is listening on. In general, that is `443`, but you
can check it via `kubectl -n <namespace> describe svc gardener-extension-provider-equinix-metal` in the soil cluster, for example:

```
$ kubectl -n extension-provider-equinix-metal-9r7xh describe svc gardener-extension-provider-equinix-metal
Name:              gardener-extension-provider-equinix-metal
Namespace:         extension-provider-equinix-metal-9r7xh
Labels:            app=gardener-extension-provider-equinix-metal
                   app.kubernetes.io/instance=provider-equinix-metal
                   app.kubernetes.io/name=gardener-extension-provider-equinix-metal
Annotations:       cloud.google.com/neg: {"ingress":true}
                   resources.gardener.cloud/description:
                     DO NOT EDIT - This resource is managed by gardener-resource-manager.
                     Any modifications are discarded and the resource is returned to the original state.
Selector:          app.kubernetes.io/instance=provider-equinix-metal,app.kubernetes.io/name=gardener-extension-provider-equinix-metal
Type:              ClusterIP
IP Families:       <none>
IP:                10.80.2.212
IPs:               <none>
Port:              <unset>  443/TCP
TargetPort:        10250/TCP
Endpoints:         10.76.3.40:10250
Session Affinity:  None
Events:            <none>
```

It is listening on `443` while sending to `10250`, so you must pass `443` to `make start`.

#### Start Extension

Start the extension, which will start both the controller and the mutating webhook:

```sh
make start IGNORE_OPERATION_ANNOTATION=false WEBHOOK_CONFIG_MODE=service EXTENSION_NAMESPACE=<namespace> WEBHOOK_CONFIG_PORT=<port> WEBHOOK_CONFIG_SERICE_PORT=<service port>
```

For example:

```sh
make start IGNORE_OPERATION_ANNOTATION=false WEBHOOK_CONFIG_MODE=service EXTENSION_NAMESPACE=extension-provider-equinix-metal-9r7xh WEBHOOK_CONFIG_PORT=8443 WEBHOOK_CONFIG_SERICE_PORT=443
```

The option `IGNORE_OPERATION_ANNOTATION=false` is critically important. The annotation is how the gardenlet communicates with
the extension controller. If it is set to `true` (the default), then shoots will not get reconciles by our extension.

#### Webhook Tunnel

```sh
./hack/hook-me.sh equinixmetal extension-provider-equinix-metal-9r7xh 8443 10250
```

##### What the Webhook Does

The Webhook mutates certain resources as they are applied to the cluster. Specifically, it focuses on configuring the elements that will
be deployed to `Shoot` clusters. These include, but are not limited to:

* `kube-controller-manager` - the one that will be started for the `Shoot`, actually running as a pod in the `Seed`, to have the correct `--cloud-provider` setting
* `OperatingSystemConfig` - to have the correct configuration for the specific OS deployed on the cloud provider's nodes

Full details on the webhooks is available [here](https://github.com/gardener/gardener/blob/master/docs/extensions/controlplane-webhooks.md).

### Likely Errors

When running the `hook-me.sh` or `make start`, here are some common errors.

#### Bad Template

`hook-me.sh` might give output as follows:

```
error: error executing template "{{ index (index  .status.loadBalancer.ingress 0).ip }}": template: output:1:10: executing "output" at <index .status.loadBalancer.ingress 0>: error calling index: index of untyped nil
```

If you already ran `make start`, just ignore it. This happens because the `hook-me.sh` is trying to get the load balancer IP
assigned to the `inlets-lb` Service, but it hasn't been assigned yet. If you wait, this should clean itself up.

#### Address Not Found

`hook-me.sh` might give output as follows:

```
host: couldn'\''t get address for '\''executing'\'': not found
```

Again, this is likely a transient error. Wait until it passes.

#### Inlets Pod Not Ready

`hook-me.sh` gives output as follows:

```
++ kubectl -n extension-provider-equinix-metal-9r7xh get pods inlets-server --no-headers
++ awk '{print $2}'
+ test 2/3 = 3/3
+ sleep 2s
```

Note that the number of ready pods given by `test 2/3` might also be `0/3` or `1/3`.

To resolve this, you need to get the status of the `inlets-server` pod:

```
kubectl -n <namespace> describe pod inlets-server
```

For example:

```
kubectl -n extension-provider-equinix-metal-9r7xh describe pod inlets-server
```

Here you could find one of several errors:

##### TLS Mount Failed

```
Type     Reason       Age                     From     Message
----     ------       ----                    ----     -------
Warning  FailedMount  28m (x5993 over 11d)    kubelet  Unable to attach or mount volumes: unmounted volumes=[inlets-tls], unattached volumes=[default-token-m8cwf inlets-tls]: timed out waiting for the condition
Warning  FailedMount  8m30s (x1614 over 11d)  kubelet  Unable to attach or mount volumes: unmounted volumes=[inlets-tls], unattached volumes=[inlets-tls default-token-m8cwf]: timed out waiting for the condition
Warning  FailedMount  4m17s (x8491 over 11d)  kubelet  MountVolume.SetUp failed for volume "inlets-tls" : secret "gardener-extension-webhook-cert" not found
```

This error likely is caused because you did not start the extension with the appropriate options, including `<namespace>`
and `WEBHOOK_CONFIG_MODE=service`. These options set up the necessary certificate secrets.

##### Inlets Not Found

```
  Warning  Failed     2m44s (x3 over 3m33s)  kubelet            Failed to pull image "inlets/inlets:2.6.3": rpc error: code = Unknown desc = Error response from daemon: pull access denied for inlets/inlets, repository does not exist or may require 'docker login': denied: requested access to the resource is denied
```

For unknown reasons, your cluster is unable to get inlets.




### Deploying an extension to a soil cluster

>NOTE: Some operations in the development workflow don't support Go modules. It is recommended to
>clone the extension's source directory into your GOPATH.

If you don't already have it, clone the extension's source repository:

```console
mkdir -p $GOPATH/src/github.com/gardener && cd $_
git clone git@github.com:gardener/gardener-extension-provider-equinix-metal.git
cd gardener-extension-provider-equinix-metal
```

Set your `KUBECONFIG` env var to point at the kubeconfig file belonging to the **garden cluster**
(not the soil cluster). For a Gardener cluster deployed using
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

The extension is in [controller-registration.yaml](./example/controller-registration.yaml).
It contains the Helm chart from [here](./charts/gardener-extension-provider-equinix-metal/values.yaml), tarred, gzipped, and then
base64-encoded. By default, it deploys one replica of the extension pod from a registry.

>NOTE: The `controller-registration.yaml` file in the `master` branch contains a reference to an
>**unreleased** container image. Either overwrite the `tag` field in the file or check out a Git
>tag before registering the extension.

If you are developing locally, you don't actually want to run one replica; you want 0. Modify the values so that it has:

```yaml
values:
  replicaCount: 0
```

This will set up all of the prerequisites but not try to deploy a controller pod, leaving you free to run `make start`

Next, register the extension:

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
docker tag 9fa4c197cf04 <your-container-registry>:testing
docker push <your-container-registry>:testing
```

To use the new image, register or re-register the extension with the Gardener API server. Note that
when re-pushing an existing Docker tag, you will likely have to temporarily set the extension's
`imagePullPolicy` field to `Always` and delete the extension's pod to force the kubelet to re-pull
the image.

### Debugging

To check the logs of the extension controller, run the following command against the **base** API
server:

```
kubectl -n extension-provider-equinix-metal-xxxxx logs gardener-extension-provider-equinix-metal-xxxxx
```

### Miscellaneous and "gotchas"

#### Helm chart packaging

The Helm chart that is used by Gardener to deploy the extension controller is packaged as a
**base64-encoded tgz archive** inline inside `example/controller-registration.yaml` under the
`chart` field. This entire yaml is generated when you run `make generate`.

To view its contents, run the following command:

```
cat example/controller-registration.yaml | grep chart | awk {'print $2'} | base64 -d | tar zxvf -
```

#### Ignored files
Some generated files are **gitignored**. To avoid building images using outdated generated data, be
sure to **always** run `make generate` before running `make docker-images` (even when there is no
visible Git diff).

#### Retrying operations

If shoot creation operations fail, you can restart them. See [here](https://github.com/gardener/gardener/blob/master/docs/usage/shoot_operations.md). For example, to retry a failed shoot creation:

```
$ kubectl -n garden-em annotate shoot em-test gardener.cloud/operation=retry
shoot.core.gardener.cloud/em-test annotated
```

Remember that the shoot is in the _garden_ cluster, not the _soil_ cluster, so use the appropriate kubeconfig.
