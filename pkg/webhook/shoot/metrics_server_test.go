// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package shoot

import (
	"context"

	extensionswebhook "github.com/gardener/gardener/extensions/pkg/webhook"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Mutator", func() {
	Describe("#mutateMetricsServerDeployment", func() {
		It("should correctly set the preferred address types", func() {
			var (
				mutator    = &mutator{}
				deployment = &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "metrics-server",
						Namespace: metav1.NamespaceSystem,
					},
					Spec: appsv1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name: "metrics-server",
										Command: []string{
											"/metrics-server",
											"--authorization-always-allow-paths=/livez,/readyz",
											"--profiling=false",
											"--cert-dir=/home/certdir",
											"--secure-port=8443",
											"--kubelet-insecure-tls",
											"--kubelet-preferred-address-types=Hostname,InternalDNS,InternalIP,ExternalDNS,ExternalIP",
											"--tls-cert-file=/srv/metrics-server/tls/tls.crt",
											"--tls-private-key-file=/srv/metrics-server/tls/tls.key",
										},
									},
								},
							},
						},
					},
				}
			)

			Expect(mutator.mutateMetricsServerDeployment(context.TODO(), deployment)).To(Succeed())

			c := extensionswebhook.ContainerWithName(deployment.Spec.Template.Spec.Containers, "metrics-server")
			Expect(c).To(Not(BeNil()))
			Expect(c.Command).To(ContainElement("--kubelet-preferred-address-types=InternalIP,ExternalIP"))
		})
	})
})
