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

package packet_test

import (
	"context"
	"errors"

	. "github.com/gardener/gardener-extension-provider-packet/pkg/packet"

	mockclient "github.com/gardener/gardener-extensions/pkg/mock/controller-runtime/client"
	kutil "github.com/gardener/gardener/pkg/utils/kubernetes"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
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

			c.EXPECT().Get(ctx, kutil.Key(namespace, name), gomock.AssignableToTypeOf(&corev1.Secret{})).DoAndReturn(func(_ context.Context, _ client.ObjectKey, secret *corev1.Secret) error {
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
