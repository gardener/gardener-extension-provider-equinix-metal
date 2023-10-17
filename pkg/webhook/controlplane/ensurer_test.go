// Copyright (c) 2019 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package controlplane

import (
	"context"
	"testing"

	"github.com/Masterminds/semver"
	"github.com/coreos/go-systemd/v22/unit"
	extensionscontroller "github.com/gardener/gardener/extensions/pkg/controller"
	extensionswebhook "github.com/gardener/gardener/extensions/pkg/webhook"
	gcontext "github.com/gardener/gardener/extensions/pkg/webhook/context"
	"github.com/gardener/gardener/extensions/pkg/webhook/controlplane/genericmutator"
	"github.com/gardener/gardener/extensions/pkg/webhook/controlplane/test"
	gardencorev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	v1beta1constants "github.com/gardener/gardener/pkg/apis/core/v1beta1/constants"
	mockclient "github.com/gardener/gardener/pkg/mock/controller-runtime/client"
	"github.com/gardener/gardener/pkg/utils/imagevector"
	testutils "github.com/gardener/gardener/pkg/utils/test"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	vpaautoscalingv1 "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	kubeletconfigv1beta1 "k8s.io/kubelet/config/v1beta1"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	namespace = "test"
)

var (
	ctx context.Context
)

func TestController(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Controlplane Webhook Suite")
}

var _ = Describe("Ensurer", func() {
	var (
		ctrl *gomock.Controller
		c    *mockclient.MockClient

		dummyContext   = gcontext.NewGardenContext(nil, nil)
		eContextK8s126 = gcontext.NewInternalGardenContext(
			&extensionscontroller.Cluster{
				Shoot: &gardencorev1beta1.Shoot{
					Spec: gardencorev1beta1.ShootSpec{
						Kubernetes: gardencorev1beta1.Kubernetes{
							Version: "1.26.0",
						},
					},
					Status: gardencorev1beta1.ShootStatus{
						TechnicalID: namespace,
					},
				},
			},
		)

		secretKey = client.ObjectKey{Namespace: namespace, Name: v1beta1constants.SecretNameCloudProvider}
		secret    = &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: v1beta1constants.SecretNameCloudProvider},
			Data:       map[string][]byte{"foo": []byte("bar")},
		}

		annotations = map[string]string{
			"checksum/secret-" + v1beta1constants.SecretNameCloudProvider: "8bafb35ff1ac60275d62e1cbd495aceb511fb354f74a20f7d06ecb48b3a68432",
		}
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		c = mockclient.NewMockClient(ctrl)
		ctx = context.TODO()
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Describe("#EnsureKubeAPIServerDeployment", func() {
		It("should add missing elements to kube-apiserver deployment", func() {
			var (
				dep = &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: v1beta1constants.DeploymentNameKubeAPIServer},
					Spec: appsv1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name: "kube-apiserver",
									},
								},
							},
						},
					},
				}
			)

			c.EXPECT().Get(ctx, secretKey, &corev1.Secret{}).DoAndReturn(clientGet(secret))

			ensurer := NewEnsurer(c, logger, false)

			// Call EnsureKubeAPIServerDeployment method and check the result
			err := ensurer.EnsureKubeAPIServerDeployment(ctx, eContextK8s126, dep, nil)
			Expect(err).To(Not(HaveOccurred()))
			checkKubeAPIServerDeployment(dep, "1.26.0", annotations)
		})

		It("should modify existing elements of kube-apiserver deployment", func() {
			var (
				dep = &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: v1beta1constants.DeploymentNameKubeAPIServer},
					Spec: appsv1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name: "kube-apiserver",
										Command: []string{
											"--enable-admission-plugins=Priority,PersistentVolumeLabel",
											"--disable-admission-plugins=",
											"--feature-gates=Foo=true",
										},
									},
								},
							},
						},
					},
				}
			)

			c.EXPECT().Get(ctx, secretKey, &corev1.Secret{}).DoAndReturn(clientGet(secret))

			ensurer := NewEnsurer(c, logger, false)

			// Call EnsureKubeAPIServerDeployment method and check the result
			err := ensurer.EnsureKubeAPIServerDeployment(ctx, eContextK8s126, dep, nil)
			Expect(err).To(Not(HaveOccurred()))
			checkKubeAPIServerDeployment(dep, "1.26.0", annotations)
		})

		It("should keep the NODE_NETWORK env variable in the kube-apiserver deployment if its value does not change", func() {
			var (
				dep = &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: v1beta1constants.DeploymentNameKubeAPIServer},
					Spec: appsv1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name: "kube-apiserver",
									},
									{
										Name: "vpn-seed",
									},
								},
							},
						},
					},
				}
				oldDep            = dep.DeepCopy()
				nodeNetworkEnvVar = corev1.EnvVar{
					Name:  "NODE_NETWORK",
					Value: "foobar",
				}
			)

			oldDep.Spec.Template.Spec.Containers[1].Env = []corev1.EnvVar{nodeNetworkEnvVar}

			c.EXPECT().Get(ctx, secretKey, &corev1.Secret{}).DoAndReturn(clientGet(secret))

			ensurer := NewEnsurer(c, logger, false)

			// Call EnsureKubeAPIServerDeployment method and check the result
			err := ensurer.EnsureKubeAPIServerDeployment(ctx, eContextK8s126, dep, oldDep)
			Expect(err).To(Not(HaveOccurred()))
			checkKubeAPIServerDeployment(dep, "1.26.0", annotations)

			c := extensionswebhook.ContainerWithName(dep.Spec.Template.Spec.Containers, "vpn-seed")
			Expect(c).To(Not(BeNil()))
			Expect(c.Env).To(ConsistOf(nodeNetworkEnvVar))
		})

		It("should not keep the NODE_NETWORK env variable in the kube-apiserver deployment if its value changes", func() {
			var (
				dep = &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: v1beta1constants.DeploymentNameKubeAPIServer},
					Spec: appsv1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name: "kube-apiserver",
									},
									{
										Name: "vpn-seed",
										Env: []corev1.EnvVar{{
											Name:  "NODE_NETWORK",
											Value: "foobar",
										}},
									},
								},
							},
						},
					},
				}
				oldDep   = dep.DeepCopy()
				newValue = "barfoo"
			)

			dep.Spec.Template.Spec.Containers[1].Env[0].Value = newValue

			c.EXPECT().Get(ctx, secretKey, &corev1.Secret{}).DoAndReturn(clientGet(secret))

			ensurer := NewEnsurer(c, logger, false)

			// Call EnsureKubeAPIServerDeployment method and check the result
			err := ensurer.EnsureKubeAPIServerDeployment(ctx, eContextK8s126, dep, oldDep)
			Expect(err).To(Not(HaveOccurred()))
			checkKubeAPIServerDeployment(dep, "1.26.0", annotations)

			c := extensionswebhook.ContainerWithName(dep.Spec.Template.Spec.Containers, "vpn-seed")
			Expect(c).To(Not(BeNil()))
			Expect(c.Env).To(ConsistOf(corev1.EnvVar{
				Name:  "NODE_NETWORK",
				Value: newValue,
			}))
		})
	})

	Describe("#EnsureKubeControllerManagerDeployment", func() {
		It("should add missing elements to kube-controller-manager deployment", func() {
			var (
				dep = &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: v1beta1constants.DeploymentNameKubeControllerManager},
					Spec: appsv1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name: "kube-controller-manager",
									},
								},
							},
						},
					},
				}
			)

			ensurer := NewEnsurer(c, logger, false)

			// Call EnsureKubeControllerManagerDeployment method and check the result
			err := ensurer.EnsureKubeControllerManagerDeployment(ctx, dummyContext, dep, nil)
			Expect(err).To(Not(HaveOccurred()))
			checkKubeControllerManagerDeployment(dep)
		})

		It("should modify existing elements of kube-controller-manager deployment", func() {
			var (
				dep = &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{Name: v1beta1constants.DeploymentNameKubeControllerManager},
					Spec: appsv1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name: "kube-controller-manager",
										Command: []string{
											"--cloud-provider=?",
										},
									},
								},
							},
						},
					},
				}
			)

			ensurer := NewEnsurer(c, logger, false)

			// Call EnsureKubeControllerManagerDeployment method and check the result
			err := ensurer.EnsureKubeControllerManagerDeployment(ctx, dummyContext, dep, nil)
			Expect(err).To(Not(HaveOccurred()))
			checkKubeControllerManagerDeployment(dep)
		})
	})

	Describe("#EnsureVPNSeedServerDeployment", func() {
		It("should keep the NODE_NETWORK env variable in the vpn-seed-server deployment if its value does not change", func() {
			var (
				dep = &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: v1beta1constants.DeploymentNameVPNSeedServer},
					Spec: appsv1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name: "vpn-seed-server",
									},
								},
							},
						},
					},
				}
				oldDep            = dep.DeepCopy()
				nodeNetworkEnvVar = corev1.EnvVar{
					Name:  "NODE_NETWORK",
					Value: "foobar",
				}
			)

			oldDep.Spec.Template.Spec.Containers[0].Env = []corev1.EnvVar{nodeNetworkEnvVar}

			ensurer := NewEnsurer(c, logger, false)

			// Call EnsureVPNSeedServerDeployment method and check the result
			err := ensurer.EnsureVPNSeedServerDeployment(ctx, dummyContext, dep, oldDep)
			Expect(err).To(Not(HaveOccurred()))

			c := extensionswebhook.ContainerWithName(dep.Spec.Template.Spec.Containers, "vpn-seed-server")
			Expect(c).To(Not(BeNil()))
			Expect(c.Env).To(ConsistOf(nodeNetworkEnvVar))
		})

		It("should not keep the NODE_NETWORK env variable in the vpn-seed-server deployment if its value changes", func() {
			var (
				dep = &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: v1beta1constants.DeploymentNameVPNSeedServer},
					Spec: appsv1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name: "vpn-seed-server",
										Env: []corev1.EnvVar{{
											Name:  "NODE_NETWORK",
											Value: "foobar",
										}},
									},
								},
							},
						},
					},
				}
				oldDep   = dep.DeepCopy()
				newValue = "barfoo"
			)

			dep.Spec.Template.Spec.Containers[0].Env[0].Value = newValue

			ensurer := NewEnsurer(c, logger, false)

			// Call EnsureKubeAPIServerDeployment method and check the result
			err := ensurer.EnsureVPNSeedServerDeployment(ctx, dummyContext, dep, oldDep)
			Expect(err).To(Not(HaveOccurred()))

			c := extensionswebhook.ContainerWithName(dep.Spec.Template.Spec.Containers, "vpn-seed-server")
			Expect(c).To(Not(BeNil()))
			Expect(c.Env).To(ConsistOf(corev1.EnvVar{
				Name:  "NODE_NETWORK",
				Value: newValue,
			}))
		})
	})

	Describe("#EnsureKubeletServiceUnitOptions", func() {
		DescribeTable("should modify existing elements of kubelet.service unit options",
			func(kubeletVersion *semver.Version, cloudProvider string) {
				var (
					oldUnitOptions = []*unit.UnitOption{
						{
							Section: "Service",
							Name:    "ExecStart",
							Value: `/opt/bin/hyperkube kubelet \
    --config=/var/lib/kubelet/config/kubelet`,
						},
					}
					newUnitOptions = []*unit.UnitOption{
						{
							Section: "Service",
							Name:    "ExecStart",
							Value: `/opt/bin/hyperkube kubelet \
    --config=/var/lib/kubelet/config/kubelet`,
						},
					}
				)

				if cloudProvider != "" {
					newUnitOptions[0].Value += ` \
    --cloud-provider=` + cloudProvider
				}

				ensurer := NewEnsurer(c, logger, false)

				// Call EnsureKubeletServiceUnitOptions method and check the result
				opts, err := ensurer.EnsureKubeletServiceUnitOptions(ctx, dummyContext, kubeletVersion, oldUnitOptions, nil)
				Expect(err).To(Not(HaveOccurred()))
				Expect(opts).To(Equal(newUnitOptions))
			},

			Entry("kubelet version >= 1.24", semver.MustParse("1.26.0"), "external"),
		)
	})

	Describe("#EnsureKubeletConfiguration", func() {
		DescribeTable("should modify existing elements of kubelet configuration",
			func(kubeletVersion *semver.Version, featureGates map[string]bool) {
				var (
					oldKubeletConfig = &kubeletconfigv1beta1.KubeletConfiguration{
						FeatureGates: map[string]bool{
							"Foo": true,
						},
					}
					newKubeletConfig = &kubeletconfigv1beta1.KubeletConfiguration{
						FeatureGates:                 featureGates,
						EnableControllerAttachDetach: pointer.Bool(true),
					}
				)
				newKubeletConfig.FeatureGates["Foo"] = true

				ensurer := NewEnsurer(c, logger, false)

				// Call EnsureKubeletConfiguration method and check the result
				kubeletConfig := *oldKubeletConfig
				err := ensurer.EnsureKubeletConfiguration(ctx, dummyContext, kubeletVersion, &kubeletConfig, nil)
				Expect(err).To(Not(HaveOccurred()))
				Expect(&kubeletConfig).To(Equal(newKubeletConfig))
			},

			Entry("kubelet version >= 1.24", semver.MustParse("1.26.0"), map[string]bool{}),
		)
	})

	Describe("#EnsureMachineControllerManagerDeployment", func() {
		var (
			ensurer    genericmutator.Ensurer
			deployment *appsv1.Deployment
		)

		BeforeEach(func() {
			deployment = &appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Namespace: "foo"}}
		})

		Context("when gardenlet does not manage MCM", func() {
			BeforeEach(func() {
				ensurer = NewEnsurer(c, logger, false)
			})

			It("should do nothing", func() {
				deploymentBefore := deployment.DeepCopy()
				Expect(ensurer.EnsureMachineControllerManagerDeployment(context.TODO(), nil, deployment, nil)).To(BeNil())
				Expect(deployment).To(Equal(deploymentBefore))
			})
		})

		Context("when gardenlet manages MCM", func() {
			BeforeEach(func() {
				ensurer = NewEnsurer(c, logger, true)
				DeferCleanup(testutils.WithVar(&ImageVector, imagevector.ImageVector{{
					Name:       "machine-controller-manager-provider-equinix-metal",
					Repository: "foo",
					Tag:        pointer.String("bar"),
				}}))
			})

			It("should inject the sidecar container", func() {
				Expect(deployment.Spec.Template.Spec.Containers).To(BeEmpty())
				Expect(ensurer.EnsureMachineControllerManagerDeployment(context.TODO(), nil, deployment, nil)).To(BeNil())
				Expect(deployment.Spec.Template.Spec.Containers).To(ConsistOf(corev1.Container{
					Name:            "machine-controller-manager-provider-equinix-metal",
					Image:           "foo:bar",
					ImagePullPolicy: corev1.PullIfNotPresent,
					Command: []string{
						"./machine-controller",
						"--control-kubeconfig=inClusterConfig",
						"--machine-creation-timeout=20m",
						"--machine-drain-timeout=2h",
						"--machine-health-timeout=10m",
						"--machine-safety-apiserver-statuscheck-timeout=30s",
						"--machine-safety-apiserver-statuscheck-period=1m",
						"--machine-safety-orphan-vms-period=30m",
						"--namespace=" + deployment.Namespace,
						"--port=10259",
						"--target-kubeconfig=/var/run/secrets/gardener.cloud/shoot/generic-kubeconfig/kubeconfig",
						"--v=3",
					},
					LivenessProbe: &corev1.Probe{
						ProbeHandler: corev1.ProbeHandler{
							HTTPGet: &corev1.HTTPGetAction{
								Path:   "/healthz",
								Port:   intstr.FromInt(10259),
								Scheme: "HTTP",
							},
						},
						InitialDelaySeconds: 30,
						TimeoutSeconds:      5,
						PeriodSeconds:       10,
						SuccessThreshold:    1,
						FailureThreshold:    3,
					},
					VolumeMounts: []corev1.VolumeMount{{
						Name:      "kubeconfig",
						MountPath: "/var/run/secrets/gardener.cloud/shoot/generic-kubeconfig",
						ReadOnly:  true,
					}},
				}))
			})
		})
	})

	Describe("#EnsureMachineControllerManagerVPA", func() {
		var (
			ensurer genericmutator.Ensurer
			vpa     *vpaautoscalingv1.VerticalPodAutoscaler
		)

		BeforeEach(func() {
			vpa = &vpaautoscalingv1.VerticalPodAutoscaler{}
		})

		Context("when gardenlet does not manage MCM", func() {
			BeforeEach(func() {
				ensurer = NewEnsurer(c, logger, false)
			})

			It("should do nothing", func() {
				vpaBefore := vpa.DeepCopy()
				Expect(ensurer.EnsureMachineControllerManagerVPA(context.TODO(), nil, vpa, nil)).To(BeNil())
				Expect(vpa).To(Equal(vpaBefore))
			})
		})

		Context("when gardenlet manages MCM", func() {
			BeforeEach(func() {
				ensurer = NewEnsurer(c, logger, true)
			})

			It("should inject the sidecar container policy", func() {
				Expect(vpa.Spec.ResourcePolicy).To(BeNil())
				Expect(ensurer.EnsureMachineControllerManagerVPA(context.TODO(), nil, vpa, nil)).To(BeNil())

				ccv := vpaautoscalingv1.ContainerControlledValuesRequestsOnly
				Expect(vpa.Spec.ResourcePolicy.ContainerPolicies).To(ConsistOf(vpaautoscalingv1.ContainerResourcePolicy{
					ContainerName:    "machine-controller-manager-provider-equinix-metal",
					ControlledValues: &ccv,
					MinAllowed:       corev1.ResourceList{},
					MaxAllowed: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("2"),
						corev1.ResourceMemory: resource.MustParse("5G"),
					},
				}))
			})
		})
	})
})

func checkKubeAPIServerDeployment(dep *appsv1.Deployment, k8sVersion string, annotations map[string]string) {

	// Check that the kube-apiserver container still exists and contains all needed command line args,
	// env vars, and volume mounts
	c := extensionswebhook.ContainerWithName(dep.Spec.Template.Spec.Containers, "kube-apiserver")
	Expect(c).To(Not(BeNil()))
	Expect(c.Command).To(Not(test.ContainElementWithPrefixContaining("--enable-admission-plugins=", "PersistentVolumeLabel", ",")))
	Expect(c.Command).To(test.ContainElementWithPrefixContaining("--disable-admission-plugins=", "PersistentVolumeLabel", ","))

	Expect(c.Command).ToNot(test.ContainElementWithPrefixContaining("--feature-gates=", "VolumeSnapshotDataSource=true", ","))

	Expect(c.Command).ToNot(test.ContainElementWithPrefixContaining("--feature-gates=", "CSINodeInfo=true", ","))
	Expect(c.Command).ToNot(test.ContainElementWithPrefixContaining("--feature-gates=", "CSIDriverRegistry=true", ","))

	// Check that the Pod template contains all needed checksum annotations
	Expect(dep.Spec.Template.Annotations).To(Equal(annotations))
}

func checkKubeControllerManagerDeployment(dep *appsv1.Deployment) {
	// Check that the kube-controller-manager container still exists and contains all needed command line args,
	// env vars, and volume mounts
	c := extensionswebhook.ContainerWithName(dep.Spec.Template.Spec.Containers, "kube-controller-manager")
	Expect(c).To(Not(BeNil()))
	Expect(c.Command).To(ContainElement("--cloud-provider=external"))
}

func clientGet(result runtime.Object) interface{} {
	return func(ctx context.Context, key client.ObjectKey, obj runtime.Object, _ ...client.GetOption) error {
		switch obj.(type) {
		case *corev1.Secret:
			*obj.(*corev1.Secret) = *result.(*corev1.Secret)
		case *corev1.ConfigMap:
			*obj.(*corev1.ConfigMap) = *result.(*corev1.ConfigMap)
		}
		return nil
	}
}
