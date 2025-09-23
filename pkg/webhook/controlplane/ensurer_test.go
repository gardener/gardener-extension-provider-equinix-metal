// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package controlplane

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/coreos/go-systemd/v22/unit"
	extensionscontroller "github.com/gardener/gardener/extensions/pkg/controller"
	extensionswebhook "github.com/gardener/gardener/extensions/pkg/webhook"
	gcontext "github.com/gardener/gardener/extensions/pkg/webhook/context"
	"github.com/gardener/gardener/extensions/pkg/webhook/controlplane/genericmutator"
	"github.com/gardener/gardener/extensions/pkg/webhook/controlplane/test"
	gardencorev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	v1beta1constants "github.com/gardener/gardener/pkg/apis/core/v1beta1/constants"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/gardener/gardener/pkg/component/nodemanagement/machinecontrollermanager"
	"github.com/gardener/gardener/pkg/utils/imagevector"
	testutils "github.com/gardener/gardener/pkg/utils/test"
	mockclient "github.com/gardener/gardener/third_party/mock/controller-runtime/client"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	vpaautoscalingv1 "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	kubeletconfigv1beta1 "k8s.io/kubelet/config/v1beta1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/gardener/gardener-extension-provider-equinix-metal/pkg/apis/equinixmetal/v1alpha1"
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
		ctrl           *gomock.Controller
		c              *mockclient.MockClient
		eContextK8s131 gcontext.GardenContext
		shoot131       *gardencorev1beta1.Shoot
		infraConfig    *v1alpha1.InfrastructureConfig

		dummyContext = gcontext.NewInternalGardenContext(
			&extensionscontroller.Cluster{
				Shoot: &gardencorev1beta1.Shoot{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "my-shoot",
						Namespace: namespace,
					},
				},
			},
		)
		extObjectKey   = client.ObjectKey{Namespace: namespace, Name: "my-shoot"}
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
		infraConfig = &v1alpha1.InfrastructureConfig{
			TypeMeta: metav1.TypeMeta{
				APIVersion: v1alpha1.SchemeGroupVersion.String(),
				Kind:       "InfrastructureConfig",
			},
		}

		shoot131 = &gardencorev1beta1.Shoot{
			Spec: gardencorev1beta1.ShootSpec{
				Kubernetes: gardencorev1beta1.Kubernetes{
					Version: "1.31.1",
				},
				Provider: gardencorev1beta1.Provider{
					InfrastructureConfig: &runtime.RawExtension{
						Raw: encode(infraConfig),
					},
				},
			},
		}

		eContextK8s131 = gcontext.NewInternalGardenContext(
			&extensionscontroller.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "shoot--project--foo",
				},
				Shoot: shoot131,
			},
		)

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

			ensurer := NewEnsurer(c, logger)

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

			ensurer := NewEnsurer(c, logger)

			// Call EnsureKubeAPIServerDeployment method and check the result
			err := ensurer.EnsureKubeAPIServerDeployment(ctx, eContextK8s126, dep, nil)
			Expect(err).To(Not(HaveOccurred()))
			checkKubeAPIServerDeployment(dep, "1.26.0", annotations)
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

			ensurer := NewEnsurer(c, logger)

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

			ensurer := NewEnsurer(c, logger)

			// Call EnsureKubeControllerManagerDeployment method and check the result
			err := ensurer.EnsureKubeControllerManagerDeployment(ctx, dummyContext, dep, nil)
			Expect(err).To(Not(HaveOccurred()))
			checkKubeControllerManagerDeployment(dep)
		})
	})

	Describe("#EnsureVPNSeedServerDeployment", func() {
		It("should keep the NODE_NETWORK env variable in the vpn-seed-server deployment if its value does not change", func() {

			var (
				nodeNetworkEnvVar = corev1.EnvVar{
					Name:  "NODE_NETWORK",
					Value: "foobar",
				}
				dep = &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: v1beta1constants.DeploymentNameVPNSeedServer},
					Spec: appsv1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name: "vpn-seed-server",
										Env:  []corev1.EnvVar{nodeNetworkEnvVar},
									},
								},
							},
						},
					},
				}
				oldDep = dep.DeepCopy()
				infra  = &extensionsv1alpha1.Infrastructure{}
			)

			c.EXPECT().Get(ctx, extObjectKey, &extensionsv1alpha1.Infrastructure{}).DoAndReturn(clientGet(infra))

			ensurer := NewEnsurer(c, logger)

			// Call EnsureVPNSeedServerDeployment method and check the result
			err := ensurer.EnsureVPNSeedServerDeployment(ctx, dummyContext, dep, oldDep)
			Expect(err).To(Not(HaveOccurred()))

			c := extensionswebhook.ContainerWithName(dep.Spec.Template.Spec.Containers, "vpn-seed-server")
			Expect(c).To(Not(BeNil()))
			Expect(c.Env).To(ConsistOf(nodeNetworkEnvVar))
		})

		It("should update the NODE_NETWORK env variable in the vpn-seed-server deployment if the network changes", func() {
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
				depKey   = client.ObjectKey{Namespace: namespace, Name: v1beta1constants.DeploymentNameVPNSeedServer}
				newValue = "127.0.0.1/32"
				infra    = &extensionsv1alpha1.Infrastructure{
					Status: extensionsv1alpha1.InfrastructureStatus{
						NodesCIDR: &newValue,
					},
				}
				newDep = &appsv1.Deployment{}
			)

			c.EXPECT().Get(ctx, extObjectKey, &extensionsv1alpha1.Infrastructure{}).DoAndReturn(clientGet(infra))
			c.EXPECT().Get(ctx, depKey, &appsv1.StatefulSet{}).Return(fmt.Errorf("dummy"))
			c.EXPECT().Get(ctx, depKey, &appsv1.Deployment{}).DoAndReturn(clientGet(dep))
			c.EXPECT().
				Patch(ctx, gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ interface{}, dep *appsv1.Deployment, _ interface{}, _ ...interface{}) error {
					newDep = dep
					return nil
				})

			ensurer := NewEnsurer(c, logger)

			// Call EnsureKubeAPIServerDeployment method and check the result
			err := ensurer.EnsureVPNSeedServerDeployment(ctx, dummyContext, dep, oldDep)
			Expect(err).To(Not(HaveOccurred()))

			c := extensionswebhook.ContainerWithName(newDep.Spec.Template.Spec.Containers, "vpn-seed-server")
			Expect(c).To(Not(BeNil()))
			Expect(c.Env).To(ConsistOf(corev1.EnvVar{
				Name:  "NODE_NETWORK",
				Value: newValue,
			}))
		})
	})

	Describe("#EnsureKubeletServiceUnitOptions", func() {
		It("should modify existing elements of kubelet.service unit options", func() {
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
    --config=/var/lib/kubelet/config/kubelet \
    --cloud-provider=external`,
					},
				}
			)

			ensurer := NewEnsurer(c, logger)

			// Call EnsureKubeletServiceUnitOptions method and check the result
			opts, err := ensurer.EnsureKubeletServiceUnitOptions(ctx, dummyContext, nil, oldUnitOptions, nil)
			Expect(err).To(Not(HaveOccurred()))
			Expect(opts).To(Equal(newUnitOptions))
		})
	})

	Describe("#EnsureKubeletConfiguration", func() {
		It("should modify existing elements of kubelet configuration", func() {
			var (
				oldKubeletConfig = &kubeletconfigv1beta1.KubeletConfiguration{
					FeatureGates: map[string]bool{
						"Foo": true,
					},
				}
				newKubeletConfig = &kubeletconfigv1beta1.KubeletConfiguration{
					FeatureGates:                 map[string]bool{},
					EnableControllerAttachDetach: ptr.To(true),
				}
			)
			newKubeletConfig.FeatureGates["Foo"] = true

			ensurer := NewEnsurer(c, logger)

			// Call EnsureKubeletConfiguration method and check the result
			kubeletConfig := *oldKubeletConfig
			err := ensurer.EnsureKubeletConfiguration(ctx, dummyContext, nil, &kubeletConfig, nil)
			Expect(err).To(Not(HaveOccurred()))
			Expect(&kubeletConfig).To(Equal(newKubeletConfig))
		})
	})

	Describe("#EnsureMachineControllerManagerDeployment", func() {
		var (
			ensurer    genericmutator.Ensurer
			deployment *appsv1.Deployment
		)

		BeforeEach(func() {
			deployment = &appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Namespace: "foo"}}

			foo := "foo"
			ensurer = NewEnsurer(c, logger)
			DeferCleanup(testutils.WithVar(&ImageVector, imagevector.ImageVector{{
				Name:       "machine-controller-manager-provider-equinix-metal",
				Repository: &foo,
				Tag:        ptr.To("bar"),
			}}))
		})

		It("should inject the sidecar container", func() {
			Expect(deployment.Spec.Template.Spec.Containers).To(BeEmpty())
			Expect(ensurer.EnsureMachineControllerManagerDeployment(context.TODO(), eContextK8s131, deployment, nil)).To(BeNil())
			expectedContainer := machinecontrollermanager.ProviderSidecarContainer(shoot131, deployment.Namespace, "provider-equinix-metal", "foo:bar")
			Expect(deployment.Spec.Template.Spec.Containers).To(ConsistOf(expectedContainer))
		})
	})

	Describe("#EnsureMachineControllerManagerVPA", func() {
		var (
			ensurer genericmutator.Ensurer
			vpa     *vpaautoscalingv1.VerticalPodAutoscaler
		)

		BeforeEach(func() {
			vpa = &vpaautoscalingv1.VerticalPodAutoscaler{}
			ensurer = NewEnsurer(c, logger)
		})

		It("should inject the sidecar container policy", func() {
			Expect(vpa.Spec.ResourcePolicy).To(BeNil())
			Expect(ensurer.EnsureMachineControllerManagerVPA(context.TODO(), nil, vpa, nil)).To(BeNil())

			ccv := vpaautoscalingv1.ContainerControlledValuesRequestsOnly
			Expect(vpa.Spec.ResourcePolicy.ContainerPolicies).To(ConsistOf(vpaautoscalingv1.ContainerResourcePolicy{
				ContainerName:    "machine-controller-manager-provider-equinix-metal",
				ControlledValues: &ccv,
			}))
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
		case *appsv1.Deployment:
			*obj.(*appsv1.Deployment) = *result.(*appsv1.Deployment)
		case *extensionsv1alpha1.Infrastructure:
			*obj.(*extensionsv1alpha1.Infrastructure) = *result.(*extensionsv1alpha1.Infrastructure)
		}
		return nil
	}
}

func encode(obj runtime.Object) []byte {
	data, _ := json.Marshal(obj)
	return data
}
