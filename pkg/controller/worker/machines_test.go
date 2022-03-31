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

package worker_test

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	api "github.com/gardener/gardener-extension-provider-equinix-metal/pkg/apis/equinixmetal"
	apiv1alpha1 "github.com/gardener/gardener-extension-provider-equinix-metal/pkg/apis/equinixmetal/v1alpha1"
	. "github.com/gardener/gardener-extension-provider-equinix-metal/pkg/controller/worker"
	"github.com/gardener/gardener-extension-provider-equinix-metal/pkg/equinixmetal"

	extensionscontroller "github.com/gardener/gardener/extensions/pkg/controller"
	"github.com/gardener/gardener/extensions/pkg/controller/common"
	"github.com/gardener/gardener/extensions/pkg/controller/worker"
	genericworkeractuator "github.com/gardener/gardener/extensions/pkg/controller/worker/genericactuator"
	gardencorev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	v1beta1constants "github.com/gardener/gardener/pkg/apis/core/v1beta1/constants"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/gardener/gardener/pkg/client/kubernetes"
	mockkubernetes "github.com/gardener/gardener/pkg/client/kubernetes/mock"
	mockclient "github.com/gardener/gardener/pkg/mock/controller-runtime/client"
	machinev1alpha1 "github.com/gardener/machine-controller-manager/pkg/apis/machine/v1alpha1"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	ctx context.Context
)

var _ = Describe("Machines", func() {
	var (
		ctrl         *gomock.Controller
		c            *mockclient.MockClient
		chartApplier *mockkubernetes.MockChartApplier
		statusWriter *mockclient.MockStatusWriter
	)

	BeforeEach(func() {
		ctx = context.TODO()
		ctrl = gomock.NewController(GinkgoT())

		c = mockclient.NewMockClient(ctrl)
		chartApplier = mockkubernetes.NewMockChartApplier(ctrl)
		statusWriter = mockclient.NewMockStatusWriter(ctrl)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Context("workerDelegate", func() {
		workerDelegate, _ := NewWorkerDelegate(common.NewClientContext(nil, nil, nil), nil, "", nil, nil)

		Describe("#MachineClassKind", func() {
			It("should return the correct kind of the machine class", func() {
				Expect(workerDelegate.MachineClassKind()).To(Equal("MachineClass"))
			})
		})

		Describe("#MachineClassList", func() {
			It("should return the correct type for the machine class list", func() {
				Expect(workerDelegate.MachineClassList()).To(Equal(&machinev1alpha1.MachineClassList{}))
			})
		})

		Describe("#GenerateMachineDeployments, #DeployMachineClasses", func() {
			var (
				namespace        string
				cloudProfileName string

				apiToken  string
				projectID string
				region    string
				facility1 string
				facility2 string

				machineImageName    string
				machineImageVersion string
				machineImage        string

				machineType string
				sshKeyID    string
				userData    = []byte("some-user-data")

				namePool1           string
				minPool1            int32
				maxPool1            int32
				maxSurgePool1       intstr.IntOrString
				maxUnavailablePool1 intstr.IntOrString

				namePool2           string
				minPool2            int32
				maxPool2            int32
				maxSurgePool2       intstr.IntOrString
				maxUnavailablePool2 intstr.IntOrString

				machineConfiguration *machinev1alpha1.MachineConfiguration

				workerPoolHash1 string
				workerPoolHash2 string

				shootVersionMajorMinor string
				shootVersion           string
				scheme                 *runtime.Scheme
				decoder                runtime.Decoder
				clusterWithoutImages   *extensionscontroller.Cluster
				cluster                *extensionscontroller.Cluster
				w                      *extensionsv1alpha1.Worker
			)

			BeforeEach(func() {
				namespace = "shoot--foobar--eqxm"
				cloudProfileName = "equinix-metal"

				apiToken = "api-token"
				projectID = "project-id"
				region = "ny"
				facility1 = "ewr1"
				facility2 = "ny5"

				machineImageName = "my-os"
				machineImageVersion = "123"
				machineImage = "uuid"

				machineType = "large"
				sshKeyID = "1-2-3-4"
				userData = []byte("some-user-data")

				namePool1 = "pool-1"
				minPool1 = 5
				maxPool1 = 10
				maxSurgePool1 = intstr.FromInt(3)
				maxUnavailablePool1 = intstr.FromInt(2)

				namePool2 = "pool-2"
				minPool2 = 30
				maxPool2 = 45
				maxSurgePool2 = intstr.FromInt(10)
				maxUnavailablePool2 = intstr.FromInt(15)

				machineConfiguration = &machinev1alpha1.MachineConfiguration{}

				shootVersionMajorMinor = "1.2"
				shootVersion = shootVersionMajorMinor + ".3"

				clusterWithoutImages = &extensionscontroller.Cluster{
					Shoot: &gardencorev1beta1.Shoot{
						Spec: gardencorev1beta1.ShootSpec{
							Kubernetes: gardencorev1beta1.Kubernetes{
								Version: shootVersion,
							},
						},
					},
				}

				cloudProfileConfig := &apiv1alpha1.CloudProfileConfig{
					TypeMeta: metav1.TypeMeta{
						APIVersion: apiv1alpha1.SchemeGroupVersion.String(),
						Kind:       "CloudProfileConfig",
					},
					MachineImages: []apiv1alpha1.MachineImages{
						{
							Name: machineImageName,
							Versions: []apiv1alpha1.MachineImageVersion{
								{
									Version: machineImageVersion,
									ID:      machineImage,
								},
							},
						},
					},
				}
				cloudProfileConfigJSON, _ := json.Marshal(cloudProfileConfig)
				cluster = &extensionscontroller.Cluster{
					CloudProfile: &gardencorev1beta1.CloudProfile{
						ObjectMeta: metav1.ObjectMeta{
							Name: cloudProfileName,
						},
						Spec: gardencorev1beta1.CloudProfileSpec{
							ProviderConfig: &runtime.RawExtension{
								Raw: cloudProfileConfigJSON,
							},
						},
					},
					Shoot: clusterWithoutImages.Shoot,
				}

				w = &extensionsv1alpha1.Worker{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: namespace,
					},
					Spec: extensionsv1alpha1.WorkerSpec{
						SecretRef: corev1.SecretReference{
							Name:      "secret",
							Namespace: namespace,
						},
						Region: region,
						InfrastructureProviderStatus: &runtime.RawExtension{
							Raw: encode(&api.InfrastructureStatus{
								SSHKeyID: sshKeyID,
							}),
						},
						Pools: []extensionsv1alpha1.WorkerPool{
							{
								Name:           namePool1,
								Minimum:        minPool1,
								Maximum:        maxPool1,
								MaxSurge:       maxSurgePool1,
								MaxUnavailable: maxUnavailablePool1,
								MachineType:    machineType,
								MachineImage: extensionsv1alpha1.MachineImage{
									Name:    machineImageName,
									Version: machineImageVersion,
								},
								UserData: userData,
								Zones: []string{
									facility1,
									facility2,
								},
							},
							{
								Name:           namePool2,
								Minimum:        minPool2,
								Maximum:        maxPool2,
								MaxSurge:       maxSurgePool2,
								MaxUnavailable: maxUnavailablePool2,
								MachineType:    machineType,
								MachineImage: extensionsv1alpha1.MachineImage{
									Name:    machineImageName,
									Version: machineImageVersion,
								},
								UserData: userData,
							},
						},
					},
				}

				scheme = runtime.NewScheme()
				_ = api.AddToScheme(scheme)
				_ = apiv1alpha1.AddToScheme(scheme)
				decoder = serializer.NewCodecFactory(scheme, serializer.EnableStrict).UniversalDecoder()

				workerPoolHash1, _ = worker.WorkerPoolHash(w.Spec.Pools[0], cluster)
				workerPoolHash2, _ = worker.WorkerPoolHash(w.Spec.Pools[1], cluster)

				workerDelegate, _ = NewWorkerDelegate(common.NewClientContext(c, scheme, decoder), chartApplier, "", w, clusterWithoutImages)
			})

			Describe("machine images", func() {
				var (
					defaultMachineClass map[string]interface{}
					machineDeployments  worker.MachineDeployments
					machineClasses      map[string]interface{}
				)

				BeforeEach(func() {
					defaultMachineClass = map[string]interface{}{
						"OS":           machineImage,
						"projectID":    projectID,
						"billingCycle": "hourly",
						"machineType":  machineType,
						"metro":        region,
						"sshKeys":      []string{sshKeyID},
						"tags": []string{
							fmt.Sprintf("kubernetes.io/cluster/%s", namespace),
							"kubernetes.io/role/node",
						},
						"secret": map[string]interface{}{
							"cloudConfig": string(userData),
						},
					}

					var (
						machineClassPool1 = copyMachineClass(defaultMachineClass)
						machineClassPool2 = copyMachineClass(defaultMachineClass)

						machineClassNamePool1 = fmt.Sprintf("%s-%s", namespace, namePool1)
						machineClassNamePool2 = fmt.Sprintf("%s-%s", namespace, namePool2)

						machineClassWithHashPool1 = fmt.Sprintf("%s-%s", machineClassNamePool1, workerPoolHash1)
						machineClassWithHashPool2 = fmt.Sprintf("%s-%s", machineClassNamePool2, workerPoolHash2)
					)

					addNameAndSecretToMachineClass(machineClassPool1, apiToken, machineClassWithHashPool1, w.Spec.SecretRef)
					addNameAndSecretToMachineClass(machineClassPool2, apiToken, machineClassWithHashPool2, w.Spec.SecretRef)

					machineClassPool1["facilities"] = []string{facility1, facility2}

					machineClasses = map[string]interface{}{"machineClasses": []map[string]interface{}{
						machineClassPool1,
						machineClassPool2,
					}}

					machineDeployments = worker.MachineDeployments{
						{
							Name:                 machineClassNamePool1,
							ClassName:            machineClassWithHashPool1,
							SecretName:           machineClassWithHashPool1,
							Minimum:              minPool1,
							Maximum:              maxPool1,
							MaxSurge:             maxSurgePool1,
							MaxUnavailable:       maxUnavailablePool1,
							MachineConfiguration: machineConfiguration,
						},
						{
							Name:                 machineClassNamePool2,
							ClassName:            machineClassWithHashPool2,
							SecretName:           machineClassWithHashPool2,
							Minimum:              minPool2,
							Maximum:              maxPool2,
							MaxSurge:             maxSurgePool2,
							MaxUnavailable:       maxUnavailablePool2,
							MachineConfiguration: machineConfiguration,
						},
					}
				})

				It("should return the expected machine deployments for profile image types", func() {
					workerDelegate, _ := NewWorkerDelegate(common.NewClientContext(c, scheme, decoder), chartApplier, "", w, cluster)

					expectGetSecretCallToWork(c, apiToken, projectID)

					// Test workerDelegate.DeployMachineClasses()

					chartApplier.
						EXPECT().
						Apply(
							ctx,
							filepath.Join(equinixmetal.InternalChartsPath, "machineclass"),
							namespace,
							"machineclass",
							kubernetes.Values(machineClasses),
						)

					err := workerDelegate.DeployMachineClasses(ctx)
					Expect(err).NotTo(HaveOccurred())

					// Test workerDelegate.UpdateMachineImagesStatus()
					expectStatusContainsMachineImages(ctx, c, statusWriter, w, []apiv1alpha1.MachineImage{
						{
							Name:    machineImageName,
							Version: machineImageVersion,
							ID:      machineImage,
						},
					})
					err = workerDelegate.UpdateMachineImagesStatus(ctx)
					Expect(err).NotTo(HaveOccurred())

					// Test workerDelegate.GenerateMachineDeployments()

					result, err := workerDelegate.GenerateMachineDeployments(ctx)
					Expect(err).NotTo(HaveOccurred())
					Expect(result).To(Equal(machineDeployments))
				})

				It("should deploy the correct machine class when using values for reserved devices", func() {
					var (
						reservationIDs      = []string{"foo", "bar"}
						reservedDevicesOnly = true
					)

					w.Spec.Pools[1].ProviderConfig = &runtime.RawExtension{Raw: encode(&api.WorkerConfig{
						ReservationIDs:      reservationIDs,
						ReservedDevicesOnly: &reservedDevicesOnly,
					})}

					newHash, err := worker.WorkerPoolHash(w.Spec.Pools[1], cluster)
					Expect(err).NotTo(HaveOccurred())

					var (
						machineClassNamePool2     = fmt.Sprintf("%s-%s", namespace, namePool2)
						machineClassWithHashPool2 = fmt.Sprintf("%s-%s", machineClassNamePool2, newHash)
					)

					machineClasses["machineClasses"].([]map[string]interface{})[1]["name"] = machineClassWithHashPool2
					machineClasses["machineClasses"].([]map[string]interface{})[1]["reservationIDs"] = reservationIDs
					machineClasses["machineClasses"].([]map[string]interface{})[1]["reservedDevicesOnly"] = reservedDevicesOnly

					expectGetSecretCallToWork(c, apiToken, projectID)

					workerDelegate, _ := NewWorkerDelegate(common.NewClientContext(c, scheme, decoder), chartApplier, "", w, cluster)

					chartApplier.
						EXPECT().
						Apply(
							ctx,
							filepath.Join(equinixmetal.InternalChartsPath, "machineclass"),
							namespace,
							"machineclass",
							kubernetes.Values(machineClasses),
						)

					Expect(workerDelegate.DeployMachineClasses(context.TODO())).NotTo(HaveOccurred())
				})
			})

			It("should fail because the secret cannot be read", func() {
				c.EXPECT().
					Get(ctx, gomock.Any(), gomock.AssignableToTypeOf(&corev1.Secret{})).
					Return(fmt.Errorf("error"))

				result, err := workerDelegate.GenerateMachineDeployments(ctx)
				Expect(err).To(HaveOccurred())
				Expect(result).To(BeNil())
			})

			It("should fail because the version is invalid", func() {
				expectGetSecretCallToWork(c, apiToken, projectID)

				clusterWithoutImages.Shoot.Spec.Kubernetes.Version = "invalid"
				workerDelegate, _ = NewWorkerDelegate(common.NewClientContext(c, scheme, decoder), chartApplier, "", w, cluster)

				result, err := workerDelegate.GenerateMachineDeployments(ctx)
				Expect(err).To(HaveOccurred())
				Expect(result).To(BeNil())
			})

			It("should return err when the infrastructure provider status cannot be decoded", func() {
				w.Spec.InfrastructureProviderStatus = &runtime.RawExtension{Raw: []byte(`invalid`)}

				workerDelegate, _ = NewWorkerDelegate(common.NewClientContext(c, scheme, decoder), chartApplier, "", w, cluster)

				result, err := workerDelegate.GenerateMachineDeployments(ctx)
				Expect(err).To(HaveOccurred())
				Expect(result).To(BeNil())
			})

			It("should fail because the machine image cannot be found", func() {
				expectGetSecretCallToWork(c, apiToken, projectID)

				workerDelegate, _ = NewWorkerDelegate(common.NewClientContext(c, scheme, decoder), chartApplier, "", w, clusterWithoutImages)

				result, err := workerDelegate.GenerateMachineDeployments(ctx)
				Expect(err).To(HaveOccurred())
				Expect(result).To(BeNil())
			})

			It("should set expected machineControllerManager settings on machine deployment", func() {
				expectGetSecretCallToWork(c, apiToken, projectID)

				testDrainTimeout := metav1.Duration{Duration: 10 * time.Minute}
				testHealthTimeout := metav1.Duration{Duration: 20 * time.Minute}
				testCreationTimeout := metav1.Duration{Duration: 30 * time.Minute}
				testMaxEvictRetries := int32(30)
				testNodeConditions := []string{"ReadonlyFilesystem", "KernelDeadlock", "DiskPressure"}
				w.Spec.Pools[0].MachineControllerManagerSettings = &gardencorev1beta1.MachineControllerManagerSettings{
					MachineDrainTimeout:    &testDrainTimeout,
					MachineCreationTimeout: &testCreationTimeout,
					MachineHealthTimeout:   &testHealthTimeout,
					MaxEvictRetries:        &testMaxEvictRetries,
					NodeConditions:         testNodeConditions,
				}

				workerDelegate, _ = NewWorkerDelegate(common.NewClientContext(c, scheme, decoder), chartApplier, "", w, cluster)

				result, err := workerDelegate.GenerateMachineDeployments(ctx)
				resultSettings := result[0].MachineConfiguration
				resultNodeConditions := strings.Join(testNodeConditions, ",")

				Expect(err).NotTo(HaveOccurred())
				Expect(resultSettings.MachineDrainTimeout).To(Equal(&testDrainTimeout))
				Expect(resultSettings.MachineCreationTimeout).To(Equal(&testCreationTimeout))
				Expect(resultSettings.MachineHealthTimeout).To(Equal(&testHealthTimeout))
				Expect(resultSettings.MaxEvictRetries).To(Equal(&testMaxEvictRetries))
				Expect(resultSettings.NodeConditions).To(Equal(&resultNodeConditions))
			})
		})
	})
})

func encode(obj runtime.Object) []byte {
	data, _ := json.Marshal(obj)
	return data
}

func expectGetSecretCallToWork(c *mockclient.MockClient, apiToken, projectID string) {
	c.EXPECT().
		Get(ctx, gomock.Any(), gomock.AssignableToTypeOf(&corev1.Secret{})).
		DoAndReturn(func(_ context.Context, _ client.ObjectKey, secret *corev1.Secret) error {
			secret.Data = map[string][]byte{
				equinixmetal.APIToken:  []byte(apiToken),
				equinixmetal.ProjectID: []byte(projectID),
			}
			return nil
		})
}

func expectStatusContainsMachineImages(ctx context.Context, c *mockclient.MockClient, statusWriter *mockclient.MockStatusWriter, worker *extensionsv1alpha1.Worker, images []apiv1alpha1.MachineImage) {
	expectedProviderStatus := &apiv1alpha1.WorkerStatus{
		TypeMeta: metav1.TypeMeta{
			APIVersion: apiv1alpha1.SchemeGroupVersion.String(),
			Kind:       "WorkerStatus",
		},
		MachineImages: images,
	}
	workerWithExpectedStatus := worker.DeepCopy()
	workerWithExpectedStatus.Status.ProviderStatus = &runtime.RawExtension{
		Object: expectedProviderStatus,
	}

	c.EXPECT().Status().Return(statusWriter)
	statusWriter.EXPECT().Patch(ctx, workerWithExpectedStatus, gomock.Any()).Return(nil)
}

func copyMachineClass(def map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(def))

	for k, v := range def {
		out[k] = v
	}

	return out
}

func addNameAndSecretToMachineClass(class map[string]interface{}, apiToken, name string, credentialsSecretRef corev1.SecretReference) {
	class["name"] = name
	class["labels"] = map[string]string{
		v1beta1constants.GardenerPurpose: genericworkeractuator.GardenPurposeMachineClass,
	}
	class["credentialsSecretRef"] = map[string]interface{}{
		"name":      credentialsSecretRef.Name,
		"namespace": credentialsSecretRef.Namespace,
	}
}
