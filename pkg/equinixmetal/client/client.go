// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"fmt"
	"strings"

	"github.com/packethost/packngo"

	"github.com/gardener/gardener-extension-provider-equinix-metal/pkg/version"
)

type eqxmClient struct {
	client *packngo.Client
}

// NewClient creates a new Client for the given Equinix Metal credentials
func NewClient(apiKey string) ClientInterface {
	token := strings.TrimSpace(apiKey)

	if token != "" {
		client := packngo.NewClientWithAuth("gardener", token, nil)
		client.UserAgent = fmt.Sprintf("gardener-extension-provider-equinix-metal/%s %s", version.Version, client.UserAgent)
		return &eqxmClient{client}
	}

	return nil
}

func (p *eqxmClient) DeviceGet(id string) (device *packngo.Device, err error) {
	device, _, err = p.client.Devices.Get(id, &packngo.GetOptions{Includes: []string{"ip_addresses.parent_block,parent_block"}})
	return device, err
}

func (p *eqxmClient) NetworkGet(id string) (addr *packngo.IPAddressReservation, err error) {
	addr, _, err = p.client.ProjectIPs.Get(id, &packngo.GetOptions{Includes: []string{"parent_block"}})
	return addr, err
}
