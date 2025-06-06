// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package imagevector

import (
	_ "embed"

	"github.com/gardener/gardener/pkg/utils/imagevector"
	"k8s.io/apimachinery/pkg/util/runtime"

	"github.com/gardener/gardener-extension-provider-equinix-metal/pkg/equinixmetal"
)

var (
	//go:embed images.yaml
	ImagesYAML  string
	imageVector imagevector.ImageVector
)

func init() {
	var err error

	imageVector, err = imagevector.Read([]byte(ImagesYAML))
	runtime.Must(err)

	imageVector, err = imagevector.WithEnvOverride(imageVector, "IMAGEVECTOR_OVERWRITE")
	runtime.Must(err)
}

// ImageVector is the image vector that contains all the needed images.
func ImageVector() imagevector.ImageVector {
	return imageVector
}

// TerraformerImage retrieves the Terraformer image.
func TerraformerImage() string {
	image, err := imageVector.FindImage(equinixmetal.TerraformerImageName)
	runtime.Must(err)
	return image.String()
}
