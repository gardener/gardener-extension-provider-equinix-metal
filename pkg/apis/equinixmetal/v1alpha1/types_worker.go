// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// WorkerConfig contains configuration settings for the worker nodes.
type WorkerConfig struct {
	metav1.TypeMeta `json:",inline"`

	// ReservationIDs is the list of IDs of reserved devices.
	// +optional
	ReservationIDs []string `json:"reservationIDs,omitempty"`
	// ReservedDevicesOnly indicates whether only reserved devices should be used (based on the list of reservation IDs) when
	// new machines are created. If false and the list of reservation IDs is exhausted then the next available device
	// (unreserved) will be used. Default: false
	// +optional.
	ReservedDevicesOnly *bool `json:"reservedDevicesOnly,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// WorkerStatus contains information about created worker resources.
type WorkerStatus struct {
	metav1.TypeMeta `json:",inline"`

	// MachineImages is a list of machine images that have been used in this worker. Usually, the extension controller
	// gets the mapping from name/version to the provider-specific machine image data in its componentconfig. However, if
	// a version that is still in use gets removed from this componentconfig it cannot reconcile anymore existing `Worker`
	// resources that are still using this version. Hence, it stores the used versions in the provider status to ensure
	// reconciliation is possible.
	// +optional
	MachineImages []MachineImage `json:"machineImages,omitempty"`
}

// MachineImage is a mapping from logical names and versions to provider-specific machine image data.
type MachineImage struct {
	// Name is the logical name of the machine image.
	Name string `json:"name"`
	// Version is the logical version of the machine image.
	Version string `json:"version"`
	// ID is the id of the image.
	// +optional
	ID string `json:"id,omitempty"`
	// IPXEScriptURL is url to point to a IPXE script.
	// +optional
	IPXEScriptURL string `json:"ipxeScriptUrl,omitempty"`
}
