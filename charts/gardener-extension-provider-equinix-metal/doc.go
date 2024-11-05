// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

//go:generate sh -c "bash $GARDENER_HACK_DIR/generate-controller-registration.sh provider-equinix-metal . $(cat ../../VERSION) ../../example/controller-registration.yaml ControlPlane:equinixmetal Infrastructure:equinixmetal Worker:equinixmetal"

// Package chart enables go:generate support for generating the correct controller registration.
package chart
