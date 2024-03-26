// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package infrastructure_test

import (
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/gardener/gardener-extension-provider-equinix-metal/pkg/controller/infrastructure"
	"github.com/gardener/gardener-extension-provider-equinix-metal/pkg/equinixmetal"
)

var _ = Describe("Actuator Reconcile", func() {
	Describe("#GenerateTerraformInfraConfig", func() {
		It("should compute the correct Terraform config", func() {
			var (
				sshKey      = "foo-bar"
				clusterName = "shoot--foo-bar"

				infrastructure = &extensionsv1alpha1.Infrastructure{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "infra",
						Namespace: clusterName,
					},
					Spec: extensionsv1alpha1.InfrastructureSpec{
						SSHPublicKey: []byte(sshKey),
					},
				}
			)

			Expect(GenerateTerraformInfraConfig(infrastructure)).To(Equal(map[string]interface{}{
				"sshPublicKey": sshKey,
				"clusterName":  clusterName,
				"outputKeys": map[string]interface{}{
					"sshKeyID": equinixmetal.SSHKeyID,
				},
			}))
		})
	})
})
