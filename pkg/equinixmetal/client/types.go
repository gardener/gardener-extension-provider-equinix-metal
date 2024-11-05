// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"context"

	"github.com/equinix/equinix-sdk-go/services/metalv1"
)

// ClientInterface is an interface which must be implemented by Equinix Metal clients.
type ClientInterface interface {
	GetDevice(
		ctx context.Context,
		deviceID string,
	) (*metalv1.Device, error)
	GetNetwork(
		ctx context.Context,
		projectID string,
	) (*metalv1.IPReservationList, error)
}
