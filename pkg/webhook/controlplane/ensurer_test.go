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
	"github.com/gardener/gardener/extensions/pkg/webhook/controlplane/test"
	gardencorev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	v1beta1constants "github.com/gardener/gardener/pkg/apis/core/v1beta1/constants"
	mockclient "github.com/gardener/gardener/pkg/mock/controller-runtime/client"
	"github.com/gardener/gardener/pkg/utils/version"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	kubeletconfigv1beta1 "k8s.io/kubelet/config/v1beta1"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/runtime/inject"
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

		dummyContext   = gcontext.NewGardenContext(nil, nil)
		eContextK8s120 = gcontext.NewInternalGardenContext(
			&extensionscontroller.Cluster{
				Shoot: &gardencorev1beta1.Shoot{
					Spec: gardencorev1beta1.ShootSpec{
						Kubernetes: gardencorev1beta1.Kubernetes{
							Version: "1.20.0",
						},
					},
					Status: gardencorev1beta1.ShootStatus{
						TechnicalID: namespace,
					},
				},
			},
		)
		eContextK8s121 = gcontext.NewInternalGardenContext(
			&extensionscontroller.Cluster{
				Shoot: &gardencorev1beta1.Shoot{
					Spec: gardencorev1beta1.ShootSpec{
						Kubernetes: gardencorev1beta1.Kubernetes{
							Version: "1.21.0",
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
		ctx = context.TODO()
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Describe("#EnsureKubeAPIServerDeployment", func() {
		It("should add missing elements to kube-apiserver deployment (k8s <= 1.20)", func() {
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

			// Create mock client
			client := mockclient.NewMockClient(ctrl)
			client.EXPECT().Get(ctx, secretKey, &corev1.Secret{}).DoAndReturn(clientGet(secret))

			// Create ensurer
			ensurer := NewEnsurer(logger)
			err := ensurer.(inject.Client).InjectClient(client)
			Expect(err).To(Not(HaveOccurred()))

			// Call EnsureKubeAPIServerDeployment method and check the result
			err = ensurer.EnsureKubeAPIServerDeployment(ctx, eContextK8s120, dep, nil)
			Expect(err).To(Not(HaveOccurred()))
			checkKubeAPIServerDeployment(dep, "1.20.0", annotations)
		})

		It("should add missing elements to kube-apiserver deployment (k8s >= 1.21)", func() {
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

			// Create mock client
			client := mockclient.NewMockClient(ctrl)
			client.EXPECT().Get(ctx, secretKey, &corev1.Secret{}).DoAndReturn(clientGet(secret))

			// Create ensurer
			ensurer := NewEnsurer(logger)
			err := ensurer.(inject.Client).InjectClient(client)
			Expect(err).To(Not(HaveOccurred()))

			// Call EnsureKubeAPIServerDeployment method and check the result
			err = ensurer.EnsureKubeAPIServerDeployment(ctx, eContextK8s121, dep, nil)
			Expect(err).To(Not(HaveOccurred()))
			checkKubeAPIServerDeployment(dep, "1.21.0", annotations)
		})

		It("should modify existing elements of kube-apiserver deployment (k8s <= 1.20)", func() {
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

			// Create mock client
			client := mockclient.NewMockClient(ctrl)
			client.EXPECT().Get(ctx, secretKey, &corev1.Secret{}).DoAndReturn(clientGet(secret))

			// Create ensurer
			ensurer := NewEnsurer(logger)
			err := ensurer.(inject.Client).InjectClient(client)
			Expect(err).To(Not(HaveOccurred()))

			// Call EnsureKubeAPIServerDeployment method and check the result
			err = ensurer.EnsureKubeAPIServerDeployment(ctx, eContextK8s120, dep, nil)
			Expect(err).To(Not(HaveOccurred()))
			checkKubeAPIServerDeployment(dep, "1.20.0", annotations)
		})

		It("should modify existing elements of kube-apiserver deployment (k8s >= 1.21)", func() {
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

			// Create mock client
			client := mockclient.NewMockClient(ctrl)
			client.EXPECT().Get(ctx, secretKey, &corev1.Secret{}).DoAndReturn(clientGet(secret))

			// Create ensurer
			ensurer := NewEnsurer(logger)
			err := ensurer.(inject.Client).InjectClient(client)
			Expect(err).To(Not(HaveOccurred()))

			// Call EnsureKubeAPIServerDeployment method and check the result
			err = ensurer.EnsureKubeAPIServerDeployment(ctx, eContextK8s121, dep, nil)
			Expect(err).To(Not(HaveOccurred()))
			checkKubeAPIServerDeployment(dep, "1.21.0", annotations)
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

			// Create mock client
			client := mockclient.NewMockClient(ctrl)
			client.EXPECT().Get(ctx, secretKey, &corev1.Secret{}).DoAndReturn(clientGet(secret))

			// Create ensurer
			ensurer := NewEnsurer(logger)
			err := ensurer.(inject.Client).InjectClient(client)
			Expect(err).To(Not(HaveOccurred()))

			// Call EnsureKubeAPIServerDeployment method and check the result
			err = ensurer.EnsureKubeAPIServerDeployment(ctx, eContextK8s120, dep, oldDep)
			Expect(err).To(Not(HaveOccurred()))
			checkKubeAPIServerDeployment(dep, "1.20.0", annotations)

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

			// Create mock client
			client := mockclient.NewMockClient(ctrl)
			client.EXPECT().Get(ctx, secretKey, &corev1.Secret{}).DoAndReturn(clientGet(secret))

			// Create ensurer
			ensurer := NewEnsurer(logger)
			err := ensurer.(inject.Client).InjectClient(client)
			Expect(err).To(Not(HaveOccurred()))

			// Call EnsureKubeAPIServerDeployment method and check the result
			err = ensurer.EnsureKubeAPIServerDeployment(ctx, eContextK8s120, dep, oldDep)
			Expect(err).To(Not(HaveOccurred()))
			checkKubeAPIServerDeployment(dep, "1.20.0", annotations)

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

			// Create ensurer
			ensurer := NewEnsurer(logger)

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

			// Create ensurer
			ensurer := NewEnsurer(logger)

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

			// Create mock client
			client := mockclient.NewMockClient(ctrl)

			// Create ensurer
			ensurer := NewEnsurer(logger)
			err := ensurer.(inject.Client).InjectClient(client)
			Expect(err).To(Not(HaveOccurred()))

			// Call EnsureVPNSeedServerDeployment method and check the result
			err = ensurer.EnsureVPNSeedServerDeployment(ctx, dummyContext, dep, oldDep)
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

			// Create mock client
			client := mockclient.NewMockClient(ctrl)

			// Create ensurer
			ensurer := NewEnsurer(logger)
			err := ensurer.(inject.Client).InjectClient(client)
			Expect(err).To(Not(HaveOccurred()))

			// Call EnsureKubeAPIServerDeployment method and check the result
			err = ensurer.EnsureVPNSeedServerDeployment(ctx, dummyContext, dep, oldDep)
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
			func(kubeletVersion *semver.Version, cloudProvider string, withControllerAttachDetachFlag bool) {
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

				if withControllerAttachDetachFlag {
					newUnitOptions[0].Value += ` \
    --enable-controller-attach-detach=true`
				}

				// Create ensurer
				ensurer := NewEnsurer(logger)

				// Call EnsureKubeletServiceUnitOptions method and check the result
				opts, err := ensurer.EnsureKubeletServiceUnitOptions(ctx, dummyContext, kubeletVersion, oldUnitOptions, nil)
				Expect(err).To(Not(HaveOccurred()))
				Expect(opts).To(Equal(newUnitOptions))
			},

			Entry("kubelet version < 1.23", semver.MustParse("1.22.0"), "external", true),
			Entry("kubelet version >= 1.23", semver.MustParse("1.23.0"), "", false),
		)
	})

	Describe("#EnsureKubeletConfiguration", func() {
		DescribeTable("should modify existing elements of kubelet configuration",
			func(kubeletVersion *semver.Version, enableControllerAttachDetach *bool) {
				var (
					oldKubeletConfig = &kubeletconfigv1beta1.KubeletConfiguration{
						FeatureGates: map[string]bool{
							"Foo": true,
						},
					}
					newKubeletConfig = &kubeletconfigv1beta1.KubeletConfiguration{
						FeatureGates: map[string]bool{
							"Foo":                      true,
							"VolumeSnapshotDataSource": true,
							"CSINodeInfo":              true,
							"CSIDriverRegistry":        true,
						},
						EnableControllerAttachDetach: enableControllerAttachDetach,
					}
				)

				// Create ensurer
				ensurer := NewEnsurer(logger)

				// Call EnsureKubeletConfiguration method and check the result
				kubeletConfig := *oldKubeletConfig
				err := ensurer.EnsureKubeletConfiguration(ctx, dummyContext, kubeletVersion, &kubeletConfig, nil)
				Expect(err).To(Not(HaveOccurred()))
				Expect(&kubeletConfig).To(Equal(newKubeletConfig))
			},

			Entry("kubelet version < 1.23", semver.MustParse("1.22.0"), nil),
			Entry("kubelet version >= 1.23", semver.MustParse("1.23.0"), pointer.Bool(true)),
		)
	})
})

func checkKubeAPIServerDeployment(dep *appsv1.Deployment, k8sVersion string, annotations map[string]string) {
	k8sVersionLowerThan121, _ := version.CompareVersions(k8sVersion, "<", "1.21")

	// Check that the kube-apiserver container still exists and contains all needed command line args,
	// env vars, and volume mounts
	c := extensionswebhook.ContainerWithName(dep.Spec.Template.Spec.Containers, "kube-apiserver")
	Expect(c).To(Not(BeNil()))
	Expect(c.Command).To(Not(test.ContainElementWithPrefixContaining("--enable-admission-plugins=", "PersistentVolumeLabel", ",")))
	Expect(c.Command).To(test.ContainElementWithPrefixContaining("--disable-admission-plugins=", "PersistentVolumeLabel", ","))
	Expect(c.Command).To(test.ContainElementWithPrefixContaining("--feature-gates=", "VolumeSnapshotDataSource=true", ","))

	if k8sVersionLowerThan121 {
		Expect(c.Command).To(test.ContainElementWithPrefixContaining("--feature-gates=", "CSINodeInfo=true", ","))
		Expect(c.Command).To(test.ContainElementWithPrefixContaining("--feature-gates=", "CSIDriverRegistry=true", ","))
	} else {
		Expect(c.Command).ToNot(test.ContainElementWithPrefixContaining("--feature-gates=", "CSINodeInfo=true", ","))
		Expect(c.Command).ToNot(test.ContainElementWithPrefixContaining("--feature-gates=", "CSIDriverRegistry=true", ","))
	}

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
	return func(ctx context.Context, key client.ObjectKey, obj runtime.Object) error {
		switch obj.(type) {
		case *corev1.Secret:
			*obj.(*corev1.Secret) = *result.(*corev1.Secret)
		case *corev1.ConfigMap:
			*obj.(*corev1.ConfigMap) = *result.(*corev1.ConfigMap)
		}
		return nil
	}
}
