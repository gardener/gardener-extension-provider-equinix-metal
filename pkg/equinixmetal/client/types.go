// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"github.com/packethost/packngo"
)

// ClientInterface is an interface which must be implemented by Equinix Metal clients.
type ClientInterface interface {
	DeviceGet(id string) (device *packngo.Device, err error)
	NetworkGet(id string) (addr *packngo.IPAddressReservation, err error)
}
