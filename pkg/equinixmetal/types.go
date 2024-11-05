// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package equinixmetal

const (
	// Name is the name of the Equinix Metal provider controller.
	Name = "provider-equinix-metal"

	// TerraformerImageName is the name of the Terraformer image.
	TerraformerImageName = "terraformer"
	// CloudControllerManagerImageName is the name of the cloud-controller-manager image.
	CloudControllerManagerImageName = "cloud-controller-manager"
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

	// CloudControllerManagerName is a constant for the name of the CloudController deployed by the worker controller.
	CloudControllerManagerName = "cloud-controller-manager"
)

// Credentials stores Equinix Metal credentials.
type Credentials struct {
	// APIToken is the API token.
	APIToken []byte
	// ProjectID is the ID of the project.
	ProjectID []byte
}
