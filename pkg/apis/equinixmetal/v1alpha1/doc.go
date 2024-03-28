// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

// +k8s:deepcopy-gen=package
// +k8s:conversion-gen=github.com/gardener/gardener-extension-provider-equinix-metal/pkg/apis/equinixmetal
// +k8s:openapi-gen=true
// +k8s:defaulter-gen=TypeMeta

//go:generate gen-crd-api-reference-docs -api-dir . -config ../../../../hack/api-reference/api.json -template-dir ../../../../vendor/github.com/gardener/gardener/hack/api-reference/template -out-file ../../../../hack/api-reference/api.md

// Package v1alpha1 contains the Equinix Metal provider API resources.
// +groupName=equinixmetal.provider.extensions.gardener.cloud
package v1alpha1 // import "github.com/gardener/gardener-extension-provider-equinix-metal/pkg/apis/equinixmetal/v1alpha1"
