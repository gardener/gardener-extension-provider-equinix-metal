#!/bin/bash
#
# SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
#
# SPDX-License-Identifier: Apache-2.0

set -o errexit
set -o nounset
set -o pipefail

# We need to explicitly pass GO111MODULE=off to k8s.io/code-generator as it is significantly slower otherwise,
# see https://github.com/kubernetes/code-generator/issues/100.
export GO111MODULE=off

rm -f $GOPATH/bin/*-gen

PROJECT_ROOT=$(dirname $0)/..

bash "${PROJECT_ROOT}"/vendor/k8s.io/code-generator/generate-internal-groups.sh \
  deepcopy,defaulter \
  github.com/gardener/gardener-extension-provider-equinix-metal/pkg/client \
  github.com/gardener/gardener-extension-provider-equinix-metal/pkg/apis \
  github.com/gardener/gardener-extension-provider-equinix-metal/pkg/apis \
  "equinixmetal:v1alpha1" \
  --go-header-file "${PROJECT_ROOT}/vendor/github.com/gardener/gardener/hack/LICENSE_BOILERPLATE.txt"

bash "${PROJECT_ROOT}"/vendor/k8s.io/code-generator/generate-internal-groups.sh \
  conversion \
  github.com/gardener/gardener-extension-provider-equinix-metal/pkg/client \
  github.com/gardener/gardener-extension-provider-equinix-metal/pkg/apis \
  github.com/gardener/gardener-extension-provider-equinix-metal/pkg/apis \
  "equinixmetal:v1alpha1" \
  --extra-peer-dirs=github.com/gardener/gardener-extension-provider-equinix-metal/pkg/apis/equinixmetal,github.com/gardener/gardener-extension-provider-equinix-metal/pkg/apis/equinixmetal/v1alpha1,k8s.io/apimachinery/pkg/apis/meta/v1,k8s.io/apimachinery/pkg/conversion,k8s.io/apimachinery/pkg/runtime \
  --go-header-file "${PROJECT_ROOT}/vendor/github.com/gardener/gardener/hack/LICENSE_BOILERPLATE.txt"


bash "${PROJECT_ROOT}"/vendor/k8s.io/code-generator/generate-internal-groups.sh \
  deepcopy,defaulter \
  github.com/gardener/gardener-extension-provider-equinix-metal/pkg/client/componentconfig \
  github.com/gardener/gardener-extension-provider-equinix-metal/pkg/apis \
  github.com/gardener/gardener-extension-provider-equinix-metal/pkg/apis \
  "config:v1alpha1" \
  --go-header-file "${PROJECT_ROOT}/vendor/github.com/gardener/gardener/hack/LICENSE_BOILERPLATE.txt"

bash "${PROJECT_ROOT}"/vendor/k8s.io/code-generator/generate-internal-groups.sh \
  conversion \
  github.com/gardener/gardener-extension-provider-equinix-metal/pkg/client/componentconfig \
  github.com/gardener/gardener-extension-provider-equinix-metal/pkg/apis \
  github.com/gardener/gardener-extension-provider-equinix-metal/pkg/apis \
  "config:v1alpha1" \
  --extra-peer-dirs=github.com/gardener/gardener-extension-provider-equinix-metal/pkg/apis/config,github.com/gardener/gardener-extension-provider-equinix-metal/pkg/apis/config/v1alpha1,k8s.io/apimachinery/pkg/apis/meta/v1,k8s.io/apimachinery/pkg/conversion,k8s.io/apimachinery/pkg/runtime,github.com/gardener/gardener/extensions/pkg/apis/config/v1alpha1 \
  --go-header-file "${PROJECT_ROOT}/vendor/github.com/gardener/gardener/hack/LICENSE_BOILERPLATE.txt"
