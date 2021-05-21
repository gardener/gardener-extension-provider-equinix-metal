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

package infrastructure

import (
	"bytes"
	"context"
	"fmt"

	packetv1alpha1 "github.com/gardener/gardener-extension-provider-packet/pkg/apis/packet/v1alpha1"
	"github.com/gardener/gardener-extension-provider-packet/pkg/packet"

	extensionscontroller "github.com/gardener/gardener/extensions/pkg/controller"
	"github.com/gardener/gardener/extensions/pkg/terraformer"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (a *actuator) Reconcile(ctx context.Context, infrastructure *extensionsv1alpha1.Infrastructure, cluster *extensionscontroller.Cluster) error {
	logger := a.logger.WithValues("infrastructure", client.ObjectKeyFromObject(infrastructure), "operation", "reconcile")
	return a.reconcile(ctx, logger, infrastructure, cluster, terraformer.StateConfigMapInitializerFunc(terraformer.CreateState))
}

func (a *actuator) reconcile(ctx context.Context, logger logr.Logger, infrastructure *extensionsv1alpha1.Infrastructure, cluster *extensionscontroller.Cluster, stateInitializer terraformer.StateConfigMapInitializer) error {
	credentials, err := packet.GetCredentialsFromSecretRef(ctx, a.Client(), infrastructure.Spec.SecretRef)
	if err != nil {
		return err
	}

	var (
		terraformConfig = GenerateTerraformInfraConfig(infrastructure, string(credentials.ProjectID))
		mainTF          bytes.Buffer
	)

	if err := tplMainTF.Execute(&mainTF, terraformConfig); err != nil {
		return fmt.Errorf("could not render Terraform template: %+v", err)
	}

	tf, err := a.newTerraformer(logger, packet.TerraformerPurposeInfra, infrastructure)
	if err != nil {
		return fmt.Errorf("could not create terraformer object: %+v", err)
	}

	if err := tf.
		SetEnvVars(generateTerraformInfraVariablesEnvironment(infrastructure.Spec.SecretRef)...).
		InitializeWith(ctx,
			terraformer.DefaultInitializer(
				a.Client(),
				mainTF.String(),
				variablesTF,
				[]byte(terraformTFVars),
				stateInitializer,
			)).
		Apply(ctx); err != nil {

		return errors.Wrap(err, "failed to apply the terraform config")
	}

	return a.updateProviderStatus(ctx, tf, infrastructure)
}

// GenerateTerraformInfraConfig generates the Packet Terraform configuration based on the given infrastructure and project.
func GenerateTerraformInfraConfig(infrastructure *extensionsv1alpha1.Infrastructure, projectID string) map[string]interface{} {
	return map[string]interface{}{
		"packet": map[string]interface{}{
			"projectID": projectID,
		},
		"sshPublicKey": string(infrastructure.Spec.SSHPublicKey),
		"clusterName":  infrastructure.Namespace,
		"outputKeys": map[string]interface{}{
			"sshKeyID": packet.SSHKeyID,
		},
	}
}

func (a *actuator) updateProviderStatus(ctx context.Context, tf terraformer.Terraformer, infrastructure *extensionsv1alpha1.Infrastructure) error {
	outputVarKeys := []string{
		packet.SSHKeyID,
	}

	output, err := tf.GetStateOutputVariables(ctx, outputVarKeys...)
	if err != nil {
		return err
	}

	state, err := tf.GetRawState(ctx)
	if err != nil {
		return err
	}
	stateByte, err := state.Marshal()
	if err != nil {
		return err
	}

	return extensionscontroller.TryUpdateStatus(ctx, retry.DefaultBackoff, a.Client(), infrastructure, func() error {
		infrastructure.Status.ProviderStatus = &runtime.RawExtension{
			Object: &packetv1alpha1.InfrastructureStatus{
				TypeMeta: metav1.TypeMeta{
					APIVersion: packetv1alpha1.SchemeGroupVersion.String(),
					Kind:       "InfrastructureStatus",
				},
				SSHKeyID: output[packet.SSHKeyID],
			},
		}
		infrastructure.Status.State = &runtime.RawExtension{Raw: stateByte}
		return nil
	})
}
