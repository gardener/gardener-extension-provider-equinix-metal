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
	"errors"
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"

	"github.com/gardener/gardener-extension-provider-equinix-metal/pkg/apis/config"
	extensionswebhook "github.com/gardener/gardener/extensions/pkg/webhook"
	resourcemanagerv1alpha1 "github.com/gardener/gardener/pkg/resourcemanager/apis/config/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	extensionscontextwebhook "github.com/gardener/gardener/extensions/pkg/webhook/context"
)

var (
	configDecoder runtime.Decoder
)

func init() {
	configScheme := runtime.NewScheme()
	schemeBuilder := runtime.NewSchemeBuilder(
		config.AddToScheme,
		resourcemanagerv1alpha1.AddToScheme,
	)
	utilruntime.Must(schemeBuilder.AddToScheme(configScheme))
	configDecoder = serializer.NewCodecFactory(configScheme).UniversalDecoder()
}

var _ extensionswebhook.Mutator = &customMutator{}

type customMutator struct {
	delegateMutator extensionswebhook.Mutator
	client          client.Client
}

func (m *customMutator) Mutate(ctx context.Context, new, old client.Object) error {
	passthrough, err := m.mutate(ctx, new, old)
	if err != nil {
		return err
	}
	if !passthrough {
		return nil
	}

	return m.delegateMutator.Mutate(ctx, new, old)
}

func (m *customMutator) mutate(ctx context.Context, new, old client.Object) (bool, error) {
	if new.GetDeletionTimestamp() != nil {
		return true, nil
	}
	gctx := extensionscontextwebhook.NewGardenContext(m.client, new)

	switch x := new.(type) {
	case *corev1.ConfigMap:
		if strings.HasPrefix(x.GetName(), "gardener-resource-manager") {
			var oldCm *corev1.ConfigMap
			if old != nil {
				var ok bool
				oldCm, ok = old.(*corev1.ConfigMap)
				if !ok {
					return false, errors.New("could not cast old object to corev1.ConfigMap")
				}
			}
			return false, m.ensureGardenerResourceManagerConfigMap(ctx, gctx, x, oldCm)
		}
	}

	return true, nil
}

func (m *customMutator) ensureGardenerResourceManagerConfigMap(
	ctx context.Context,
	gctx extensionscontextwebhook.GardenContext,
	new, old *corev1.ConfigMap,
) error {
	logger.V(1).Info("Mutate resource manager config")

	config := &resourcemanagerv1alpha1.ResourceManagerConfiguration{}
	if err := runtime.DecodeInto(configDecoder, []byte(new.Data["config.yaml"]), config); err != nil {
		return fmt.Errorf("error decoding config: %w", err)
	}

	config.TargetClientConnection.Namespaces = append(config.TargetClientConnection.Namespaces, "metallb-system")

	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}
	new.Data["config.yaml"] = string(data)
	return nil
}
