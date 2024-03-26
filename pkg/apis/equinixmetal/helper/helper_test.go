// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package helper_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	api "github.com/gardener/gardener-extension-provider-equinix-metal/pkg/apis/equinixmetal"
	. "github.com/gardener/gardener-extension-provider-equinix-metal/pkg/apis/equinixmetal/helper"
)

const configImage = "some-uuid"
const profileImage = "some-other-uuid"

var _ = Describe("Helper", func() {
	DescribeTable("#FindMachineImage",
		func(machineImages []api.MachineImage, name, version string, expectedMachineImage *api.MachineImage, expectErr bool) {
			machineImage, err := FindMachineImage(machineImages, name, version)
			expectResults(machineImage, expectedMachineImage, err, expectErr)
		},

		Entry("list is nil", nil, "foo", "1.2.3", nil, true),
		Entry("empty list", []api.MachineImage{}, "foo", "1.2.3", nil, true),
		Entry("entry not found (no name)", []api.MachineImage{{Name: "bar", Version: "1.2.3", ID: "1234"}}, "foo", "1.2.3", nil, true),
		Entry("entry not found (no version)", []api.MachineImage{{Name: "bar", Version: "1.2.3", ID: "1234"}}, "foo", "1.2.4", nil, true),
		Entry("entry exists", []api.MachineImage{{Name: "bar", Version: "1.2.3", ID: "1234"}}, "bar", "1.2.3", &api.MachineImage{Name: "bar", Version: "1.2.3", ID: "1234"}, false),
	)

	DescribeTable("#FindImage",
		func(profileImages []api.MachineImages, imageName, version string, expectedImage string) {
			cfg := &api.CloudProfileConfig{}
			cfg.MachineImages = profileImages
			image, err := FindImageFromCloudProfile(cfg, imageName, version)

			if expectedImage != "" {
				Expect(image.ID).To(Equal(expectedImage))
				Expect(err).NotTo(HaveOccurred())
			} else {
				Expect(err).To(HaveOccurred())
			}
		},

		Entry("list is nil", nil, "ubuntu", "1", ""),

		Entry("profile empty list", []api.MachineImages{}, "ubuntu", "1", ""),
		Entry("profile entry not found (image does not exist)", makeProfileMachineImages("debian", "1"), "ubuntu", "1", ""),
		Entry("profile entry not found (version does not exist)", makeProfileMachineImages("ubuntu", "2"), "ubuntu", "1", ""),
		Entry("profile entry", makeProfileMachineImages("ubuntu", "1"), "ubuntu", "1", profileImage),
	)
})

func makeProfileMachineImages(name, version string) []api.MachineImages {
	var versions []api.MachineImageVersion
	if len(configImage) != 0 {
		versions = append(versions, api.MachineImageVersion{
			Version: version,
			ID:      profileImage,
		})
	}

	return []api.MachineImages{
		{
			Name:     name,
			Versions: versions,
		},
	}
}

func expectResults(result, expected interface{}, err error, expectErr bool) {
	if !expectErr {
		Expect(result).To(Equal(expected))
		Expect(err).NotTo(HaveOccurred())
	} else {
		Expect(result).To(BeNil())
		Expect(err).To(HaveOccurred())
	}
}
