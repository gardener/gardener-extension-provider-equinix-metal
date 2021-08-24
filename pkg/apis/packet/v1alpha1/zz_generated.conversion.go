// +build !ignore_autogenerated

/*
Copyright (c) 2021 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by conversion-gen. DO NOT EDIT.

package v1alpha1

import (
	unsafe "unsafe"

	packet "github.com/gardener/gardener-extension-provider-packet/pkg/apis/packet"
	conversion "k8s.io/apimachinery/pkg/conversion"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

func init() {
	localSchemeBuilder.Register(RegisterConversions)
}

// RegisterConversions adds conversion functions to the given scheme.
// Public to allow building arbitrary schemes.
func RegisterConversions(s *runtime.Scheme) error {
	if err := s.AddGeneratedConversionFunc((*CloudProfileConfig)(nil), (*packet.CloudProfileConfig)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1alpha1_CloudProfileConfig_To_packet_CloudProfileConfig(a.(*CloudProfileConfig), b.(*packet.CloudProfileConfig), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*packet.CloudProfileConfig)(nil), (*CloudProfileConfig)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_packet_CloudProfileConfig_To_v1alpha1_CloudProfileConfig(a.(*packet.CloudProfileConfig), b.(*CloudProfileConfig), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*ControlPlaneConfig)(nil), (*packet.ControlPlaneConfig)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1alpha1_ControlPlaneConfig_To_packet_ControlPlaneConfig(a.(*ControlPlaneConfig), b.(*packet.ControlPlaneConfig), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*packet.ControlPlaneConfig)(nil), (*ControlPlaneConfig)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_packet_ControlPlaneConfig_To_v1alpha1_ControlPlaneConfig(a.(*packet.ControlPlaneConfig), b.(*ControlPlaneConfig), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*InfrastructureConfig)(nil), (*packet.InfrastructureConfig)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1alpha1_InfrastructureConfig_To_packet_InfrastructureConfig(a.(*InfrastructureConfig), b.(*packet.InfrastructureConfig), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*packet.InfrastructureConfig)(nil), (*InfrastructureConfig)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_packet_InfrastructureConfig_To_v1alpha1_InfrastructureConfig(a.(*packet.InfrastructureConfig), b.(*InfrastructureConfig), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*InfrastructureStatus)(nil), (*packet.InfrastructureStatus)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1alpha1_InfrastructureStatus_To_packet_InfrastructureStatus(a.(*InfrastructureStatus), b.(*packet.InfrastructureStatus), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*packet.InfrastructureStatus)(nil), (*InfrastructureStatus)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_packet_InfrastructureStatus_To_v1alpha1_InfrastructureStatus(a.(*packet.InfrastructureStatus), b.(*InfrastructureStatus), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*MachineImage)(nil), (*packet.MachineImage)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1alpha1_MachineImage_To_packet_MachineImage(a.(*MachineImage), b.(*packet.MachineImage), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*packet.MachineImage)(nil), (*MachineImage)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_packet_MachineImage_To_v1alpha1_MachineImage(a.(*packet.MachineImage), b.(*MachineImage), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*MachineImageVersion)(nil), (*packet.MachineImageVersion)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1alpha1_MachineImageVersion_To_packet_MachineImageVersion(a.(*MachineImageVersion), b.(*packet.MachineImageVersion), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*packet.MachineImageVersion)(nil), (*MachineImageVersion)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_packet_MachineImageVersion_To_v1alpha1_MachineImageVersion(a.(*packet.MachineImageVersion), b.(*MachineImageVersion), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*MachineImages)(nil), (*packet.MachineImages)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1alpha1_MachineImages_To_packet_MachineImages(a.(*MachineImages), b.(*packet.MachineImages), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*packet.MachineImages)(nil), (*MachineImages)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_packet_MachineImages_To_v1alpha1_MachineImages(a.(*packet.MachineImages), b.(*MachineImages), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*WorkerConfig)(nil), (*packet.WorkerConfig)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1alpha1_WorkerConfig_To_packet_WorkerConfig(a.(*WorkerConfig), b.(*packet.WorkerConfig), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*packet.WorkerConfig)(nil), (*WorkerConfig)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_packet_WorkerConfig_To_v1alpha1_WorkerConfig(a.(*packet.WorkerConfig), b.(*WorkerConfig), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*WorkerStatus)(nil), (*packet.WorkerStatus)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1alpha1_WorkerStatus_To_packet_WorkerStatus(a.(*WorkerStatus), b.(*packet.WorkerStatus), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*packet.WorkerStatus)(nil), (*WorkerStatus)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_packet_WorkerStatus_To_v1alpha1_WorkerStatus(a.(*packet.WorkerStatus), b.(*WorkerStatus), scope)
	}); err != nil {
		return err
	}
	return nil
}

func autoConvert_v1alpha1_CloudProfileConfig_To_packet_CloudProfileConfig(in *CloudProfileConfig, out *packet.CloudProfileConfig, s conversion.Scope) error {
	out.MachineImages = *(*[]packet.MachineImages)(unsafe.Pointer(&in.MachineImages))
	return nil
}

// Convert_v1alpha1_CloudProfileConfig_To_packet_CloudProfileConfig is an autogenerated conversion function.
func Convert_v1alpha1_CloudProfileConfig_To_packet_CloudProfileConfig(in *CloudProfileConfig, out *packet.CloudProfileConfig, s conversion.Scope) error {
	return autoConvert_v1alpha1_CloudProfileConfig_To_packet_CloudProfileConfig(in, out, s)
}

func autoConvert_packet_CloudProfileConfig_To_v1alpha1_CloudProfileConfig(in *packet.CloudProfileConfig, out *CloudProfileConfig, s conversion.Scope) error {
	out.MachineImages = *(*[]MachineImages)(unsafe.Pointer(&in.MachineImages))
	return nil
}

// Convert_packet_CloudProfileConfig_To_v1alpha1_CloudProfileConfig is an autogenerated conversion function.
func Convert_packet_CloudProfileConfig_To_v1alpha1_CloudProfileConfig(in *packet.CloudProfileConfig, out *CloudProfileConfig, s conversion.Scope) error {
	return autoConvert_packet_CloudProfileConfig_To_v1alpha1_CloudProfileConfig(in, out, s)
}

func autoConvert_v1alpha1_ControlPlaneConfig_To_packet_ControlPlaneConfig(in *ControlPlaneConfig, out *packet.ControlPlaneConfig, s conversion.Scope) error {
	return nil
}

// Convert_v1alpha1_ControlPlaneConfig_To_packet_ControlPlaneConfig is an autogenerated conversion function.
func Convert_v1alpha1_ControlPlaneConfig_To_packet_ControlPlaneConfig(in *ControlPlaneConfig, out *packet.ControlPlaneConfig, s conversion.Scope) error {
	return autoConvert_v1alpha1_ControlPlaneConfig_To_packet_ControlPlaneConfig(in, out, s)
}

func autoConvert_packet_ControlPlaneConfig_To_v1alpha1_ControlPlaneConfig(in *packet.ControlPlaneConfig, out *ControlPlaneConfig, s conversion.Scope) error {
	return nil
}

// Convert_packet_ControlPlaneConfig_To_v1alpha1_ControlPlaneConfig is an autogenerated conversion function.
func Convert_packet_ControlPlaneConfig_To_v1alpha1_ControlPlaneConfig(in *packet.ControlPlaneConfig, out *ControlPlaneConfig, s conversion.Scope) error {
	return autoConvert_packet_ControlPlaneConfig_To_v1alpha1_ControlPlaneConfig(in, out, s)
}

func autoConvert_v1alpha1_InfrastructureConfig_To_packet_InfrastructureConfig(in *InfrastructureConfig, out *packet.InfrastructureConfig, s conversion.Scope) error {
	return nil
}

// Convert_v1alpha1_InfrastructureConfig_To_packet_InfrastructureConfig is an autogenerated conversion function.
func Convert_v1alpha1_InfrastructureConfig_To_packet_InfrastructureConfig(in *InfrastructureConfig, out *packet.InfrastructureConfig, s conversion.Scope) error {
	return autoConvert_v1alpha1_InfrastructureConfig_To_packet_InfrastructureConfig(in, out, s)
}

func autoConvert_packet_InfrastructureConfig_To_v1alpha1_InfrastructureConfig(in *packet.InfrastructureConfig, out *InfrastructureConfig, s conversion.Scope) error {
	return nil
}

// Convert_packet_InfrastructureConfig_To_v1alpha1_InfrastructureConfig is an autogenerated conversion function.
func Convert_packet_InfrastructureConfig_To_v1alpha1_InfrastructureConfig(in *packet.InfrastructureConfig, out *InfrastructureConfig, s conversion.Scope) error {
	return autoConvert_packet_InfrastructureConfig_To_v1alpha1_InfrastructureConfig(in, out, s)
}

func autoConvert_v1alpha1_InfrastructureStatus_To_packet_InfrastructureStatus(in *InfrastructureStatus, out *packet.InfrastructureStatus, s conversion.Scope) error {
	out.SSHKeyID = in.SSHKeyID
	return nil
}

// Convert_v1alpha1_InfrastructureStatus_To_packet_InfrastructureStatus is an autogenerated conversion function.
func Convert_v1alpha1_InfrastructureStatus_To_packet_InfrastructureStatus(in *InfrastructureStatus, out *packet.InfrastructureStatus, s conversion.Scope) error {
	return autoConvert_v1alpha1_InfrastructureStatus_To_packet_InfrastructureStatus(in, out, s)
}

func autoConvert_packet_InfrastructureStatus_To_v1alpha1_InfrastructureStatus(in *packet.InfrastructureStatus, out *InfrastructureStatus, s conversion.Scope) error {
	out.SSHKeyID = in.SSHKeyID
	return nil
}

// Convert_packet_InfrastructureStatus_To_v1alpha1_InfrastructureStatus is an autogenerated conversion function.
func Convert_packet_InfrastructureStatus_To_v1alpha1_InfrastructureStatus(in *packet.InfrastructureStatus, out *InfrastructureStatus, s conversion.Scope) error {
	return autoConvert_packet_InfrastructureStatus_To_v1alpha1_InfrastructureStatus(in, out, s)
}

func autoConvert_v1alpha1_MachineImage_To_packet_MachineImage(in *MachineImage, out *packet.MachineImage, s conversion.Scope) error {
	out.Name = in.Name
	out.Version = in.Version
	out.ID = in.ID
	return nil
}

// Convert_v1alpha1_MachineImage_To_packet_MachineImage is an autogenerated conversion function.
func Convert_v1alpha1_MachineImage_To_packet_MachineImage(in *MachineImage, out *packet.MachineImage, s conversion.Scope) error {
	return autoConvert_v1alpha1_MachineImage_To_packet_MachineImage(in, out, s)
}

func autoConvert_packet_MachineImage_To_v1alpha1_MachineImage(in *packet.MachineImage, out *MachineImage, s conversion.Scope) error {
	out.Name = in.Name
	out.Version = in.Version
	out.ID = in.ID
	return nil
}

// Convert_packet_MachineImage_To_v1alpha1_MachineImage is an autogenerated conversion function.
func Convert_packet_MachineImage_To_v1alpha1_MachineImage(in *packet.MachineImage, out *MachineImage, s conversion.Scope) error {
	return autoConvert_packet_MachineImage_To_v1alpha1_MachineImage(in, out, s)
}

func autoConvert_v1alpha1_MachineImageVersion_To_packet_MachineImageVersion(in *MachineImageVersion, out *packet.MachineImageVersion, s conversion.Scope) error {
	out.Version = in.Version
	out.ID = in.ID
	return nil
}

// Convert_v1alpha1_MachineImageVersion_To_packet_MachineImageVersion is an autogenerated conversion function.
func Convert_v1alpha1_MachineImageVersion_To_packet_MachineImageVersion(in *MachineImageVersion, out *packet.MachineImageVersion, s conversion.Scope) error {
	return autoConvert_v1alpha1_MachineImageVersion_To_packet_MachineImageVersion(in, out, s)
}

func autoConvert_packet_MachineImageVersion_To_v1alpha1_MachineImageVersion(in *packet.MachineImageVersion, out *MachineImageVersion, s conversion.Scope) error {
	out.Version = in.Version
	out.ID = in.ID
	return nil
}

// Convert_packet_MachineImageVersion_To_v1alpha1_MachineImageVersion is an autogenerated conversion function.
func Convert_packet_MachineImageVersion_To_v1alpha1_MachineImageVersion(in *packet.MachineImageVersion, out *MachineImageVersion, s conversion.Scope) error {
	return autoConvert_packet_MachineImageVersion_To_v1alpha1_MachineImageVersion(in, out, s)
}

func autoConvert_v1alpha1_MachineImages_To_packet_MachineImages(in *MachineImages, out *packet.MachineImages, s conversion.Scope) error {
	out.Name = in.Name
	out.Versions = *(*[]packet.MachineImageVersion)(unsafe.Pointer(&in.Versions))
	return nil
}

// Convert_v1alpha1_MachineImages_To_packet_MachineImages is an autogenerated conversion function.
func Convert_v1alpha1_MachineImages_To_packet_MachineImages(in *MachineImages, out *packet.MachineImages, s conversion.Scope) error {
	return autoConvert_v1alpha1_MachineImages_To_packet_MachineImages(in, out, s)
}

func autoConvert_packet_MachineImages_To_v1alpha1_MachineImages(in *packet.MachineImages, out *MachineImages, s conversion.Scope) error {
	out.Name = in.Name
	out.Versions = *(*[]MachineImageVersion)(unsafe.Pointer(&in.Versions))
	return nil
}

// Convert_packet_MachineImages_To_v1alpha1_MachineImages is an autogenerated conversion function.
func Convert_packet_MachineImages_To_v1alpha1_MachineImages(in *packet.MachineImages, out *MachineImages, s conversion.Scope) error {
	return autoConvert_packet_MachineImages_To_v1alpha1_MachineImages(in, out, s)
}

func autoConvert_v1alpha1_WorkerConfig_To_packet_WorkerConfig(in *WorkerConfig, out *packet.WorkerConfig, s conversion.Scope) error {
	out.ReservationIDs = *(*[]string)(unsafe.Pointer(&in.ReservationIDs))
	out.OnlyReserved = (*bool)(unsafe.Pointer(in.OnlyReserved))
	return nil
}

// Convert_v1alpha1_WorkerConfig_To_packet_WorkerConfig is an autogenerated conversion function.
func Convert_v1alpha1_WorkerConfig_To_packet_WorkerConfig(in *WorkerConfig, out *packet.WorkerConfig, s conversion.Scope) error {
	return autoConvert_v1alpha1_WorkerConfig_To_packet_WorkerConfig(in, out, s)
}

func autoConvert_packet_WorkerConfig_To_v1alpha1_WorkerConfig(in *packet.WorkerConfig, out *WorkerConfig, s conversion.Scope) error {
	out.ReservationIDs = *(*[]string)(unsafe.Pointer(&in.ReservationIDs))
	out.OnlyReserved = (*bool)(unsafe.Pointer(in.OnlyReserved))
	return nil
}

// Convert_packet_WorkerConfig_To_v1alpha1_WorkerConfig is an autogenerated conversion function.
func Convert_packet_WorkerConfig_To_v1alpha1_WorkerConfig(in *packet.WorkerConfig, out *WorkerConfig, s conversion.Scope) error {
	return autoConvert_packet_WorkerConfig_To_v1alpha1_WorkerConfig(in, out, s)
}

func autoConvert_v1alpha1_WorkerStatus_To_packet_WorkerStatus(in *WorkerStatus, out *packet.WorkerStatus, s conversion.Scope) error {
	out.MachineImages = *(*[]packet.MachineImage)(unsafe.Pointer(&in.MachineImages))
	return nil
}

// Convert_v1alpha1_WorkerStatus_To_packet_WorkerStatus is an autogenerated conversion function.
func Convert_v1alpha1_WorkerStatus_To_packet_WorkerStatus(in *WorkerStatus, out *packet.WorkerStatus, s conversion.Scope) error {
	return autoConvert_v1alpha1_WorkerStatus_To_packet_WorkerStatus(in, out, s)
}

func autoConvert_packet_WorkerStatus_To_v1alpha1_WorkerStatus(in *packet.WorkerStatus, out *WorkerStatus, s conversion.Scope) error {
	out.MachineImages = *(*[]MachineImage)(unsafe.Pointer(&in.MachineImages))
	return nil
}

// Convert_packet_WorkerStatus_To_v1alpha1_WorkerStatus is an autogenerated conversion function.
func Convert_packet_WorkerStatus_To_v1alpha1_WorkerStatus(in *packet.WorkerStatus, out *WorkerStatus, s conversion.Scope) error {
	return autoConvert_packet_WorkerStatus_To_v1alpha1_WorkerStatus(in, out, s)
}
