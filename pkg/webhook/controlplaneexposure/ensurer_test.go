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

package controlplaneexposure

import (
	"context"
	"testing"

	druidv1alpha1 "github.com/gardener/etcd-druid/api/v1alpha1"
	extensionswebhook "github.com/gardener/gardener/extensions/pkg/webhook"
	gcontext "github.com/gardener/gardener/extensions/pkg/webhook/context"
	"github.com/gardener/gardener/extensions/pkg/webhook/controlplane/test"
	v1beta1constants "github.com/gardener/gardener/pkg/apis/core/v1beta1/constants"
	mockclient "github.com/gardener/gardener/pkg/mock/controller-runtime/client"
	"github.com/gardener/gardener/pkg/utils"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/runtime/inject"

	"github.com/gardener/gardener-extension-provider-equinix-metal/pkg/apis/config"
)

const (
	namespace    = "test"
	testIP       = "1.1.1.1"
	testHostname = "foo"
)

var (
	ctx context.Context
)

func TestController(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Controlplane Exposure Webhook Suite")
}

var _ = Describe("Ensurer", func() {
	var (
		etcdStorage = &config.ETCDStorage{
			ClassName: pointer.StringPtr("gardener.cloud-fast"),
			Capacity:  utils.QuantityPtr(resource.MustParse("25Gi")),
		}

		ctrl *gomock.Controller

		dummyContext = gcontext.NewGardenContext(nil, nil)

		serviceKey = client.ObjectKey{Namespace: namespace, Name: v1beta1constants.DeploymentNameKubeAPIServer}

		service = &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: v1beta1constants.DeploymentNameKubeAPIServer},
			Status: corev1.ServiceStatus{
				LoadBalancer: corev1.LoadBalancerStatus{
					Ingress: []corev1.LoadBalancerIngress{
						{
							Hostname: testHostname,
							IP:       testIP,
						},
					},
				},
			},
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
		It("should not modify kube-apiserver deployment if SNI is enabled", func() {
			var (
				dep = &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      v1beta1constants.DeploymentNameKubeAPIServer,
						Namespace: namespace,
						Labels:    map[string]string{"core.gardener.cloud/apiserver-exposure": "gardener-managed"},
					},
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
				depCopy = dep.DeepCopy()
			)

			// Create ensurer
			ensurer := NewEnsurer(etcdStorage, logger)

			err := ensurer.EnsureKubeAPIServerDeployment(context.TODO(), dummyContext, dep, nil)
			Expect(err).To(Not(HaveOccurred()))

			Expect(dep).To(Equal(depCopy))
		})

		It("should add missing elements to kube-apiserver deployment", func() {
			var (
				dep = &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{Name: v1beta1constants.DeploymentNameKubeAPIServer, Namespace: namespace},
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
			client.EXPECT().Get(ctx, serviceKey, &corev1.Service{}).DoAndReturn(clientGet(service))

			// Create ensurer
			ensurer := NewEnsurer(etcdStorage, logger)
			Expect(ensurer.(inject.Client).InjectClient(client)).To(Not(HaveOccurred()))

			// Call EnsureKubeAPIServerDeployment method and check the result
			err := ensurer.EnsureKubeAPIServerDeployment(context.TODO(), dummyContext, dep, nil)
			Expect(err).To(Not(HaveOccurred()))
			checkKubeAPIServerDeployment(dep)
		})

		It("should modify existing elements of kube-apiserver deployment", func() {
			var (
				dep = &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{Name: v1beta1constants.DeploymentNameKubeAPIServer, Namespace: namespace},
					Spec: appsv1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name:    "kube-apiserver",
										Command: []string{},
									},
								},
							},
						},
					},
				}
			)

			// Create mock client
			client := mockclient.NewMockClient(ctrl)
			client.EXPECT().Get(ctx, serviceKey, &corev1.Service{}).DoAndReturn(clientGet(service))

			// Create ensurer
			ensurer := NewEnsurer(etcdStorage, logger)
			Expect(ensurer.(inject.Client).InjectClient(client)).To(Not(HaveOccurred()))

			// Call EnsureKubeAPIServerDeployment method and check the result
			err := ensurer.EnsureKubeAPIServerDeployment(context.TODO(), dummyContext, dep, nil)
			Expect(err).To(Not(HaveOccurred()))
			checkKubeAPIServerDeployment(dep)
		})
	})

	Describe("#EnsureETCD", func() {
		It("should add or modify elements to etcd-main statefulset", func() {
			var (
				etcd = &druidv1alpha1.Etcd{
					ObjectMeta: metav1.ObjectMeta{Name: v1beta1constants.ETCDMain},
				}
			)

			// Create ensurer
			ensurer := NewEnsurer(etcdStorage, logger)

			// Call EnsureETCDStatefulSet method and check the result
			err := ensurer.EnsureETCD(context.TODO(), dummyContext, etcd, nil)
			Expect(err).To(Not(HaveOccurred()))
			checkETCDMain(etcd)
		})

		It("should modify existing elements of etcd-main statefulset", func() {
			var (
				r    = resource.MustParse("10Gi")
				etcd = &druidv1alpha1.Etcd{
					ObjectMeta: metav1.ObjectMeta{Name: v1beta1constants.ETCDMain},
					Spec: druidv1alpha1.EtcdSpec{
						StorageCapacity: &r,
					},
				}
			)

			// Create ensurer
			ensurer := NewEnsurer(etcdStorage, logger)

			// Call EnsureETCDStatefulSet method and check the result
			err := ensurer.EnsureETCD(context.TODO(), dummyContext, etcd, nil)
			Expect(err).To(Not(HaveOccurred()))
			checkETCDMain(etcd)
		})

		It("should add or modify elements to etcd-events statefulset", func() {
			var (
				etcd = &druidv1alpha1.Etcd{
					ObjectMeta: metav1.ObjectMeta{Name: v1beta1constants.ETCDEvents},
				}
			)

			// Create ensurer
			ensurer := NewEnsurer(etcdStorage, logger)

			// Call EnsureETCDStatefulSet method and check the result
			err := ensurer.EnsureETCD(context.TODO(), dummyContext, etcd, nil)
			Expect(err).To(Not(HaveOccurred()))
			checkETCDEvents(etcd)
		})

		It("should modify existing elements of etcd-events statefulset", func() {
			var (
				r    = resource.MustParse("20Gi")
				etcd = &druidv1alpha1.Etcd{
					ObjectMeta: metav1.ObjectMeta{Name: v1beta1constants.ETCDEvents},
					Spec: druidv1alpha1.EtcdSpec{
						StorageCapacity: &r,
					},
				}
			)

			// Create ensurer
			ensurer := NewEnsurer(etcdStorage, logger)

			// Call EnsureETCDStatefulSet method and check the result
			err := ensurer.EnsureETCD(context.TODO(), dummyContext, etcd, nil)
			Expect(err).To(Not(HaveOccurred()))
			checkETCDEvents(etcd)
		})
	})
})

func checkKubeAPIServerDeployment(dep *appsv1.Deployment) {
	// Check that the kube-apiserver container still exists and contains all needed command line args
	c := extensionswebhook.ContainerWithName(dep.Spec.Template.Spec.Containers, "kube-apiserver")
	Expect(c).To(Not(BeNil()))

	Expect(c.Command).To(test.ContainElementWithPrefixContaining("--advertise-address=", testIP, ","))
}

func checkETCDMain(etcd *druidv1alpha1.Etcd) {
	Expect(*etcd.Spec.StorageClass).To(Equal("gardener.cloud-fast"))
	Expect(*etcd.Spec.StorageCapacity).To(Equal(resource.MustParse("25Gi")))
}

func checkETCDEvents(etcd *druidv1alpha1.Etcd) {
	Expect(*etcd.Spec.StorageClass).To(Equal(""))
	Expect(*etcd.Spec.StorageCapacity).To(Equal(resource.MustParse("10Gi")))
}

func clientGet(result runtime.Object) interface{} {
	return func(ctx context.Context, key client.ObjectKey, obj runtime.Object, _ ...client.GetOption) error {
		switch obj.(type) {
		case *corev1.Service:
			*obj.(*corev1.Service) = *result.(*corev1.Service)
		}
		return nil
	}
}
