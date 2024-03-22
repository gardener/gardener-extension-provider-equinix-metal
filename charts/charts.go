// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package charts

import "embed"

//go:embed internal
var InternalChart embed.FS

const (
	InternalChartsPath = "internal"
)
