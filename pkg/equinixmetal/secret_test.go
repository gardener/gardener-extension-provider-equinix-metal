// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package equinixmetal_test

import (
	"context"
	"errors"

	mockclient "github.com/gardener/gardener/pkg/mock/controller-runtime/client"
	kutil "github.com/gardener/gardener/pkg/utils/kubernetes"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	. "github.com/gardener/gardener-extension-provider-equinix-metal/pkg/equinixmetal"
)

var _ = Describe("Secret", func() {
	Describe("#GetCredentialsFromSecretRef", func() {
		var (
			ctrl *gomock.Controller
			c    *mockclient.MockClient

			ctx       = context.TODO()
			namespace = "namespace"
			name      = "name"

			secretRef = corev1.SecretReference{
				Name:      name,
				Namespace: namespace,
			}
		)

		BeforeEach(func() {
			ctrl = gomock.NewController(GinkgoT())

			c = mockclient.NewMockClient(ctrl)
		})

		AfterEach(func() {
			ctrl.Finish()
		})

		It("should return an error because secret could not be read", func() {
			fakeErr := errors.New("error")

			c.EXPECT().Get(ctx, kutil.Key(namespace, name), gomock.AssignableToTypeOf(&corev1.Secret{})).Return(fakeErr)

			credentials, err := GetCredentialsFromSecretRef(ctx, c, secretRef)

			Expect(credentials).To(BeNil())
			Expect(err).To(Equal(fakeErr))
		})

		It("should return the correct credentials object", func() {
			var (
				apiToken  = []byte("foo")
				projectID = []byte("bar")
			)

			c.EXPECT().Get(ctx, kutil.Key(namespace, name), gomock.AssignableToTypeOf(&corev1.Secret{})).DoAndReturn(func(_ context.Context, _ client.ObjectKey, secret *corev1.Secret, _ ...client.GetOption) error {
				secret.Data = map[string][]byte{
					APIToken:  apiToken,
					ProjectID: projectID,
				}
				return nil
			})

			credentials, err := GetCredentialsFromSecretRef(ctx, c, secretRef)

			Expect(credentials).To(Equal(&Credentials{
				APIToken:  apiToken,
				ProjectID: projectID,
			}))
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("#ReadCredentialsSecret", func() {

		var secret *corev1.Secret

		BeforeEach(func() {
			secret = &corev1.Secret{}
		})

		It("should return an error because api token is missing", func() {
			credentials, err := ReadCredentialsSecret(secret)

			Expect(credentials).To(BeNil())
			Expect(err).To(HaveOccurred())
		})

		It("should return an error because project id is missing", func() {
			secret.Data = map[string][]byte{
				APIToken: []byte("foo"),
			}

			credentials, err := ReadCredentialsSecret(secret)

			Expect(credentials).To(BeNil())
			Expect(err).To(HaveOccurred())
		})

		It("should return the credentials structure", func() {
			var (
				apiToken  = []byte("foo")
				projectID = []byte("bar")
			)

			secret.Data = map[string][]byte{
				APIToken:  apiToken,
				ProjectID: projectID,
			}

			credentials, err := ReadCredentialsSecret(secret)

			Expect(credentials).To(Equal(&Credentials{
				APIToken:  apiToken,
				ProjectID: projectID,
			}))
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
