// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package validation

import (
	"fmt"

	"k8s.io/apimachinery/pkg/util/validation/field"

	api "github.com/gardener/gardener-extension-provider-equinix-metal/pkg/apis/equinixmetal"
)

// ValidateCloudProfileConfig validates a CloudProfileConfig object.
func ValidateCloudProfileConfig(cloudProfile *api.CloudProfileConfig) field.ErrorList {
	allErrs := field.ErrorList{}

	machineImagesPath := field.NewPath("machineImages")
	if len(cloudProfile.MachineImages) == 0 {
		allErrs = append(allErrs, field.Required(machineImagesPath, "must provide at least one machine image"))
	}
	for i, machineImage := range cloudProfile.MachineImages {
		idxPath := machineImagesPath.Index(i)

		if len(machineImage.Name) == 0 {
			allErrs = append(allErrs, field.Required(idxPath.Child("name"), "must provide a name"))
		}

		if len(machineImage.Versions) == 0 {
			allErrs = append(allErrs, field.Required(idxPath.Child("versions"), fmt.Sprintf("must provide at least one version for machine image %q", machineImage.Name)))
		}
		for j, version := range machineImage.Versions {
			jdxPath := idxPath.Child("versions").Index(j)

			if len(version.Version) == 0 {
				allErrs = append(allErrs, field.Required(jdxPath.Child("version"), "must provide a version"))
			}
			if len(version.ID) == 0 {
				allErrs = append(allErrs, field.Required(jdxPath.Child("id"), "must provide an id"))
			}
		}
	}

	return allErrs
}
