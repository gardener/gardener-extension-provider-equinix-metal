// Copyright (c) 2018 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
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

package client

import (
	"strings"

	"github.com/packethost/packngo"
)

type packetClient struct {
	packet *packngo.Client
}

// NewClient creates a new Client for the given Packet credentials
func NewClient(packetAPIKey string) ClientInterface {
	token := strings.TrimSpace(packetAPIKey)

	if token != "" {
		return &packetClient{packngo.NewClientWithAuth("gardener", token, nil)}
	}

	return nil
}

func (p *packetClient) DeviceGet(id string) (device *packngo.Device, err error) {
	device, _, err = p.packet.Devices.Get(id, &packngo.GetOptions{Includes: []string{"ip_addresses.parent_block,parent_block"}})
	return device, err
}

func (p *packetClient) NetworkGet(id string) (addr *packngo.IPAddressReservation, err error) {
	addr, _, err = p.packet.ProjectIPs.Get(id, &packngo.GetOptions{Includes: []string{"parent_block"}})
	return addr, err
}
