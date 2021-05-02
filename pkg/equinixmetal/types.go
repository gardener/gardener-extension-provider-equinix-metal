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

package equinixmetal

import "path/filepath"

const (
	// Name is the name of the Equinix Metal provider controller.
	Name = "provider-equinix-metal"

	// TerraformerImageName is the name of the Terraformer image.
	TerraformerImageName = "terraformer"
	// CloudControllerManagerImageName is the name of the cloud-controller-manager image.
	CloudControllerManagerImageName = "cloud-controller-manager"
	// MachineControllerManagerImageName is the name of the MachineControllerManager image.
	MachineControllerManagerImageName = "machine-controller-manager"
	// MachineControllerManagerEquinixMetalImageName is the name of the MachineControllerManager EquinixMetal providerimage.
	MachineControllerManagerEquinixMetalImageName = "machine-controller-manager-provider-equinix-metal"
	// MetabotImageName is the name of the metabot image.
	MetabotImageName = "metabot"
	// MetalLBControllerImageName is the name of the MetalLB controller image.
	MetalLBControllerImageName = "metallb-controller"
	// MetalLBSpeakerImageName is the name of the MetalLB speaker image.
	MetalLBSpeakerImageName = "metallb-speaker"

	// APIToken is a constant for the key in a cloud provider secret and backup secret that holds the Equinix Metal API token.
	APIToken = "apiToken"
	// ProjectID is a constant for the key in a cloud provider secret and backup secret that holds the Equinix Metal project id.
	ProjectID = "projectID"

	// TerraformerPurposeInfra is a constant for the complete Terraform setup with purpose 'infrastructure'.
	TerraformerPurposeInfra = "infra"
	// SSHKeyID key for accessing SSH key ID from outputs in terraform
	SSHKeyID = "key_pair_id"

	// MachineControllerManagerName is a constant for the name of the machine-controller-manager.
	MachineControllerManagerName = "machine-controller-manager"
	// MachineControllerManagerVpaName is the name of the VerticalPodAutoscaler of the machine-controller-manager deployment.
	MachineControllerManagerVpaName = "machine-controller-manager-vpa"
	// MachineControllerManagerMonitoringConfigName is the name of the ConfigMap containing monitoring stack configurations for machine-controller-manager.
	MachineControllerManagerMonitoringConfigName = "machine-controller-manager-monitoring-config"
	// CloudControllerManagerName is a constant for the name of the CloudController deployed by the worker controller.
	CloudControllerManagerName = "cloud-controller-manager"
	// RookCephImageName is the name of the Rook/Ceph image.
	RookCephImageName = "rook-ceph"
)

var (
	// ChartsPath is the path to the charts
	ChartsPath = filepath.Join("charts")
	// InternalChartsPath is the path to the internal charts
	InternalChartsPath = filepath.Join(ChartsPath, "internal")
)

// Credentials stores Equinix Metal credentials.
type Credentials struct {
	// APIToken is the API token.
	APIToken []byte
	// ProjectID is the ID of the project.
	ProjectID []byte
}
