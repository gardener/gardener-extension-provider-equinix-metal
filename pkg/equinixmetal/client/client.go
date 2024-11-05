// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/equinix/equinix-sdk-go/services/metalv1"

	"github.com/gardener/gardener-extension-provider-equinix-metal/pkg/version"
)

type eqxmClient struct {
	client *metalv1.APIClient
}

// NewClient creates a new Client for the given Equinix Metal credentials
func NewClient(apiKey string) (ClientInterface, error) {
	token := strings.TrimSpace(apiKey)

	if token == "" {
		return nil, errors.New("Equinix Metal api token required")
	}

	config := metalv1.NewConfiguration()
	config.Debug = false
	config.AddDefaultHeader("X-Auth-Token", token)
	config.UserAgent = fmt.Sprintf("gardener-extension-provider-equinix-metal/%s %s", version.Version, config.UserAgent)
	client := metalv1.NewAPIClient(config)

	return &eqxmClient{client}, nil
}

func (p *eqxmClient) GetDevice(
	ctx context.Context,
	deviceID string,
) (*metalv1.Device, error) {
	device, _, err := p.client.DevicesApi.
		FindDeviceById(ctx, deviceID).
		Include([]string{"ip_addresses.parent_block,parent_block"}).
		Execute()
	return device, err
}

func (p *eqxmClient) GetNetwork(
	ctx context.Context,
	projectID string,
) (*metalv1.IPReservationList, error) {
	addr, _, err := p.client.IPAddressesApi.
		FindIPReservations(ctx, projectID).
		Include([]string{"parent_block"}).
		Execute()
	return addr, err
}
