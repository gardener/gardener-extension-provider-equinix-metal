//go:build !ignore_autogenerated
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

	equinixmetal "github.com/gardener/gardener-extension-provider-equinix-metal/pkg/apis/equinixmetal"
	conversion "k8s.io/apimachinery/pkg/conversion"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

func init() {
	localSchemeBuilder.Register(RegisterConversions)
}

// RegisterConversions adds conversion functions to the given scheme.
// Public to allow building arbitrary schemes.
func RegisterConversions(s *runtime.Scheme) error {
	if err := s.AddGeneratedConversionFunc((*CloudProfileConfig)(nil), (*equinixmetal.CloudProfileConfig)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1alpha1_CloudProfileConfig_To_equinixmetal_CloudProfileConfig(a.(*CloudProfileConfig), b.(*equinixmetal.CloudProfileConfig), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*equinixmetal.CloudProfileConfig)(nil), (*CloudProfileConfig)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_equinixmetal_CloudProfileConfig_To_v1alpha1_CloudProfileConfig(a.(*equinixmetal.CloudProfileConfig), b.(*CloudProfileConfig), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*ControlPlaneConfig)(nil), (*equinixmetal.ControlPlaneConfig)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1alpha1_ControlPlaneConfig_To_equinixmetal_ControlPlaneConfig(a.(*ControlPlaneConfig), b.(*equinixmetal.ControlPlaneConfig), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*equinixmetal.ControlPlaneConfig)(nil), (*ControlPlaneConfig)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_equinixmetal_ControlPlaneConfig_To_v1alpha1_ControlPlaneConfig(a.(*equinixmetal.ControlPlaneConfig), b.(*ControlPlaneConfig), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*InfrastructureConfig)(nil), (*equinixmetal.InfrastructureConfig)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1alpha1_InfrastructureConfig_To_equinixmetal_InfrastructureConfig(a.(*InfrastructureConfig), b.(*equinixmetal.InfrastructureConfig), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*equinixmetal.InfrastructureConfig)(nil), (*InfrastructureConfig)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_equinixmetal_InfrastructureConfig_To_v1alpha1_InfrastructureConfig(a.(*equinixmetal.InfrastructureConfig), b.(*InfrastructureConfig), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*InfrastructureStatus)(nil), (*equinixmetal.InfrastructureStatus)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1alpha1_InfrastructureStatus_To_equinixmetal_InfrastructureStatus(a.(*InfrastructureStatus), b.(*equinixmetal.InfrastructureStatus), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*equinixmetal.InfrastructureStatus)(nil), (*InfrastructureStatus)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_equinixmetal_InfrastructureStatus_To_v1alpha1_InfrastructureStatus(a.(*equinixmetal.InfrastructureStatus), b.(*InfrastructureStatus), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*MachineImage)(nil), (*equinixmetal.MachineImage)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1alpha1_MachineImage_To_equinixmetal_MachineImage(a.(*MachineImage), b.(*equinixmetal.MachineImage), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*equinixmetal.MachineImage)(nil), (*MachineImage)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_equinixmetal_MachineImage_To_v1alpha1_MachineImage(a.(*equinixmetal.MachineImage), b.(*MachineImage), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*MachineImageVersion)(nil), (*equinixmetal.MachineImageVersion)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1alpha1_MachineImageVersion_To_equinixmetal_MachineImageVersion(a.(*MachineImageVersion), b.(*equinixmetal.MachineImageVersion), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*equinixmetal.MachineImageVersion)(nil), (*MachineImageVersion)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_equinixmetal_MachineImageVersion_To_v1alpha1_MachineImageVersion(a.(*equinixmetal.MachineImageVersion), b.(*MachineImageVersion), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*MachineImages)(nil), (*equinixmetal.MachineImages)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1alpha1_MachineImages_To_equinixmetal_MachineImages(a.(*MachineImages), b.(*equinixmetal.MachineImages), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*equinixmetal.MachineImages)(nil), (*MachineImages)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_equinixmetal_MachineImages_To_v1alpha1_MachineImages(a.(*equinixmetal.MachineImages), b.(*MachineImages), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*WorkerConfig)(nil), (*equinixmetal.WorkerConfig)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1alpha1_WorkerConfig_To_equinixmetal_WorkerConfig(a.(*WorkerConfig), b.(*equinixmetal.WorkerConfig), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*equinixmetal.WorkerConfig)(nil), (*WorkerConfig)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_equinixmetal_WorkerConfig_To_v1alpha1_WorkerConfig(a.(*equinixmetal.WorkerConfig), b.(*WorkerConfig), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*WorkerStatus)(nil), (*equinixmetal.WorkerStatus)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1alpha1_WorkerStatus_To_equinixmetal_WorkerStatus(a.(*WorkerStatus), b.(*equinixmetal.WorkerStatus), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*equinixmetal.WorkerStatus)(nil), (*WorkerStatus)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_equinixmetal_WorkerStatus_To_v1alpha1_WorkerStatus(a.(*equinixmetal.WorkerStatus), b.(*WorkerStatus), scope)
	}); err != nil {
		return err
	}
	return nil
}

func autoConvert_v1alpha1_CloudProfileConfig_To_equinixmetal_CloudProfileConfig(in *CloudProfileConfig, out *equinixmetal.CloudProfileConfig, s conversion.Scope) error {
	out.MachineImages = *(*[]equinixmetal.MachineImages)(unsafe.Pointer(&in.MachineImages))
	return nil
}

// Convert_v1alpha1_CloudProfileConfig_To_equinixmetal_CloudProfileConfig is an autogenerated conversion function.
func Convert_v1alpha1_CloudProfileConfig_To_equinixmetal_CloudProfileConfig(in *CloudProfileConfig, out *equinixmetal.CloudProfileConfig, s conversion.Scope) error {
	return autoConvert_v1alpha1_CloudProfileConfig_To_equinixmetal_CloudProfileConfig(in, out, s)
}

func autoConvert_equinixmetal_CloudProfileConfig_To_v1alpha1_CloudProfileConfig(in *equinixmetal.CloudProfileConfig, out *CloudProfileConfig, s conversion.Scope) error {
	out.MachineImages = *(*[]MachineImages)(unsafe.Pointer(&in.MachineImages))
	return nil
}

// Convert_equinixmetal_CloudProfileConfig_To_v1alpha1_CloudProfileConfig is an autogenerated conversion function.
func Convert_equinixmetal_CloudProfileConfig_To_v1alpha1_CloudProfileConfig(in *equinixmetal.CloudProfileConfig, out *CloudProfileConfig, s conversion.Scope) error {
	return autoConvert_equinixmetal_CloudProfileConfig_To_v1alpha1_CloudProfileConfig(in, out, s)
}

func autoConvert_v1alpha1_ControlPlaneConfig_To_equinixmetal_ControlPlaneConfig(in *ControlPlaneConfig, out *equinixmetal.ControlPlaneConfig, s conversion.Scope) error {
	return nil
}

// Convert_v1alpha1_ControlPlaneConfig_To_equinixmetal_ControlPlaneConfig is an autogenerated conversion function.
func Convert_v1alpha1_ControlPlaneConfig_To_equinixmetal_ControlPlaneConfig(in *ControlPlaneConfig, out *equinixmetal.ControlPlaneConfig, s conversion.Scope) error {
	return autoConvert_v1alpha1_ControlPlaneConfig_To_equinixmetal_ControlPlaneConfig(in, out, s)
}

func autoConvert_equinixmetal_ControlPlaneConfig_To_v1alpha1_ControlPlaneConfig(in *equinixmetal.ControlPlaneConfig, out *ControlPlaneConfig, s conversion.Scope) error {
	return nil
}

// Convert_equinixmetal_ControlPlaneConfig_To_v1alpha1_ControlPlaneConfig is an autogenerated conversion function.
func Convert_equinixmetal_ControlPlaneConfig_To_v1alpha1_ControlPlaneConfig(in *equinixmetal.ControlPlaneConfig, out *ControlPlaneConfig, s conversion.Scope) error {
	return autoConvert_equinixmetal_ControlPlaneConfig_To_v1alpha1_ControlPlaneConfig(in, out, s)
}

func autoConvert_v1alpha1_InfrastructureConfig_To_equinixmetal_InfrastructureConfig(in *InfrastructureConfig, out *equinixmetal.InfrastructureConfig, s conversion.Scope) error {
	return nil
}

// Convert_v1alpha1_InfrastructureConfig_To_equinixmetal_InfrastructureConfig is an autogenerated conversion function.
func Convert_v1alpha1_InfrastructureConfig_To_equinixmetal_InfrastructureConfig(in *InfrastructureConfig, out *equinixmetal.InfrastructureConfig, s conversion.Scope) error {
	return autoConvert_v1alpha1_InfrastructureConfig_To_equinixmetal_InfrastructureConfig(in, out, s)
}

func autoConvert_equinixmetal_InfrastructureConfig_To_v1alpha1_InfrastructureConfig(in *equinixmetal.InfrastructureConfig, out *InfrastructureConfig, s conversion.Scope) error {
	return nil
}

// Convert_equinixmetal_InfrastructureConfig_To_v1alpha1_InfrastructureConfig is an autogenerated conversion function.
func Convert_equinixmetal_InfrastructureConfig_To_v1alpha1_InfrastructureConfig(in *equinixmetal.InfrastructureConfig, out *InfrastructureConfig, s conversion.Scope) error {
	return autoConvert_equinixmetal_InfrastructureConfig_To_v1alpha1_InfrastructureConfig(in, out, s)
}

func autoConvert_v1alpha1_InfrastructureStatus_To_equinixmetal_InfrastructureStatus(in *InfrastructureStatus, out *equinixmetal.InfrastructureStatus, s conversion.Scope) error {
	out.SSHKeyID = in.SSHKeyID
	return nil
}

// Convert_v1alpha1_InfrastructureStatus_To_equinixmetal_InfrastructureStatus is an autogenerated conversion function.
func Convert_v1alpha1_InfrastructureStatus_To_equinixmetal_InfrastructureStatus(in *InfrastructureStatus, out *equinixmetal.InfrastructureStatus, s conversion.Scope) error {
	return autoConvert_v1alpha1_InfrastructureStatus_To_equinixmetal_InfrastructureStatus(in, out, s)
}

func autoConvert_equinixmetal_InfrastructureStatus_To_v1alpha1_InfrastructureStatus(in *equinixmetal.InfrastructureStatus, out *InfrastructureStatus, s conversion.Scope) error {
	out.SSHKeyID = in.SSHKeyID
	return nil
}

// Convert_equinixmetal_InfrastructureStatus_To_v1alpha1_InfrastructureStatus is an autogenerated conversion function.
func Convert_equinixmetal_InfrastructureStatus_To_v1alpha1_InfrastructureStatus(in *equinixmetal.InfrastructureStatus, out *InfrastructureStatus, s conversion.Scope) error {
	return autoConvert_equinixmetal_InfrastructureStatus_To_v1alpha1_InfrastructureStatus(in, out, s)
}

func autoConvert_v1alpha1_MachineImage_To_equinixmetal_MachineImage(in *MachineImage, out *equinixmetal.MachineImage, s conversion.Scope) error {
	out.Name = in.Name
	out.Version = in.Version
	out.ID = in.ID
	return nil
}

// Convert_v1alpha1_MachineImage_To_equinixmetal_MachineImage is an autogenerated conversion function.
func Convert_v1alpha1_MachineImage_To_equinixmetal_MachineImage(in *MachineImage, out *equinixmetal.MachineImage, s conversion.Scope) error {
	return autoConvert_v1alpha1_MachineImage_To_equinixmetal_MachineImage(in, out, s)
}

func autoConvert_equinixmetal_MachineImage_To_v1alpha1_MachineImage(in *equinixmetal.MachineImage, out *MachineImage, s conversion.Scope) error {
	out.Name = in.Name
	out.Version = in.Version
	out.ID = in.ID
	return nil
}

// Convert_equinixmetal_MachineImage_To_v1alpha1_MachineImage is an autogenerated conversion function.
func Convert_equinixmetal_MachineImage_To_v1alpha1_MachineImage(in *equinixmetal.MachineImage, out *MachineImage, s conversion.Scope) error {
	return autoConvert_equinixmetal_MachineImage_To_v1alpha1_MachineImage(in, out, s)
}

func autoConvert_v1alpha1_MachineImageVersion_To_equinixmetal_MachineImageVersion(in *MachineImageVersion, out *equinixmetal.MachineImageVersion, s conversion.Scope) error {
	out.Version = in.Version
	out.ID = in.ID
	return nil
}

// Convert_v1alpha1_MachineImageVersion_To_equinixmetal_MachineImageVersion is an autogenerated conversion function.
func Convert_v1alpha1_MachineImageVersion_To_equinixmetal_MachineImageVersion(in *MachineImageVersion, out *equinixmetal.MachineImageVersion, s conversion.Scope) error {
	return autoConvert_v1alpha1_MachineImageVersion_To_equinixmetal_MachineImageVersion(in, out, s)
}

func autoConvert_equinixmetal_MachineImageVersion_To_v1alpha1_MachineImageVersion(in *equinixmetal.MachineImageVersion, out *MachineImageVersion, s conversion.Scope) error {
	out.Version = in.Version
	out.ID = in.ID
	return nil
}

// Convert_equinixmetal_MachineImageVersion_To_v1alpha1_MachineImageVersion is an autogenerated conversion function.
func Convert_equinixmetal_MachineImageVersion_To_v1alpha1_MachineImageVersion(in *equinixmetal.MachineImageVersion, out *MachineImageVersion, s conversion.Scope) error {
	return autoConvert_equinixmetal_MachineImageVersion_To_v1alpha1_MachineImageVersion(in, out, s)
}

func autoConvert_v1alpha1_MachineImages_To_equinixmetal_MachineImages(in *MachineImages, out *equinixmetal.MachineImages, s conversion.Scope) error {
	out.Name = in.Name
	out.Versions = *(*[]equinixmetal.MachineImageVersion)(unsafe.Pointer(&in.Versions))
	return nil
}

// Convert_v1alpha1_MachineImages_To_equinixmetal_MachineImages is an autogenerated conversion function.
func Convert_v1alpha1_MachineImages_To_equinixmetal_MachineImages(in *MachineImages, out *equinixmetal.MachineImages, s conversion.Scope) error {
	return autoConvert_v1alpha1_MachineImages_To_equinixmetal_MachineImages(in, out, s)
}

func autoConvert_equinixmetal_MachineImages_To_v1alpha1_MachineImages(in *equinixmetal.MachineImages, out *MachineImages, s conversion.Scope) error {
	out.Name = in.Name
	out.Versions = *(*[]MachineImageVersion)(unsafe.Pointer(&in.Versions))
	return nil
}

// Convert_equinixmetal_MachineImages_To_v1alpha1_MachineImages is an autogenerated conversion function.
func Convert_equinixmetal_MachineImages_To_v1alpha1_MachineImages(in *equinixmetal.MachineImages, out *MachineImages, s conversion.Scope) error {
	return autoConvert_equinixmetal_MachineImages_To_v1alpha1_MachineImages(in, out, s)
}

func autoConvert_v1alpha1_WorkerConfig_To_equinixmetal_WorkerConfig(in *WorkerConfig, out *equinixmetal.WorkerConfig, s conversion.Scope) error {
	out.ReservationIDs = *(*[]string)(unsafe.Pointer(&in.ReservationIDs))
	out.ReservedDevicesOnly = (*bool)(unsafe.Pointer(in.ReservedDevicesOnly))
	return nil
}

// Convert_v1alpha1_WorkerConfig_To_equinixmetal_WorkerConfig is an autogenerated conversion function.
func Convert_v1alpha1_WorkerConfig_To_equinixmetal_WorkerConfig(in *WorkerConfig, out *equinixmetal.WorkerConfig, s conversion.Scope) error {
	return autoConvert_v1alpha1_WorkerConfig_To_equinixmetal_WorkerConfig(in, out, s)
}

func autoConvert_equinixmetal_WorkerConfig_To_v1alpha1_WorkerConfig(in *equinixmetal.WorkerConfig, out *WorkerConfig, s conversion.Scope) error {
	out.ReservationIDs = *(*[]string)(unsafe.Pointer(&in.ReservationIDs))
	out.ReservedDevicesOnly = (*bool)(unsafe.Pointer(in.ReservedDevicesOnly))
	return nil
}

// Convert_equinixmetal_WorkerConfig_To_v1alpha1_WorkerConfig is an autogenerated conversion function.
func Convert_equinixmetal_WorkerConfig_To_v1alpha1_WorkerConfig(in *equinixmetal.WorkerConfig, out *WorkerConfig, s conversion.Scope) error {
	return autoConvert_equinixmetal_WorkerConfig_To_v1alpha1_WorkerConfig(in, out, s)
}

func autoConvert_v1alpha1_WorkerStatus_To_equinixmetal_WorkerStatus(in *WorkerStatus, out *equinixmetal.WorkerStatus, s conversion.Scope) error {
	out.MachineImages = *(*[]equinixmetal.MachineImage)(unsafe.Pointer(&in.MachineImages))
	return nil
}

// Convert_v1alpha1_WorkerStatus_To_equinixmetal_WorkerStatus is an autogenerated conversion function.
func Convert_v1alpha1_WorkerStatus_To_equinixmetal_WorkerStatus(in *WorkerStatus, out *equinixmetal.WorkerStatus, s conversion.Scope) error {
	return autoConvert_v1alpha1_WorkerStatus_To_equinixmetal_WorkerStatus(in, out, s)
}

func autoConvert_equinixmetal_WorkerStatus_To_v1alpha1_WorkerStatus(in *equinixmetal.WorkerStatus, out *WorkerStatus, s conversion.Scope) error {
	out.MachineImages = *(*[]MachineImage)(unsafe.Pointer(&in.MachineImages))
	return nil
}

// Convert_equinixmetal_WorkerStatus_To_v1alpha1_WorkerStatus is an autogenerated conversion function.
func Convert_equinixmetal_WorkerStatus_To_v1alpha1_WorkerStatus(in *equinixmetal.WorkerStatus, out *WorkerStatus, s conversion.Scope) error {
	return autoConvert_equinixmetal_WorkerStatus_To_v1alpha1_WorkerStatus(in, out, s)
}