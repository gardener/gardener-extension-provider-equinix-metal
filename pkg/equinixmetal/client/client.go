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
