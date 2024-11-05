/*
Metal API

Contact: support@equinixmetal.com
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package metalv1

import (
	"encoding/json"
	"fmt"
)

// InterconnectionRedundancy Either 'primary', meaning a single interconnection, or 'redundant', meaning a redundant interconnection.
type InterconnectionRedundancy string

// List of Interconnection_redundancy
const (
	INTERCONNECTIONREDUNDANCY_PRIMARY   InterconnectionRedundancy = "primary"
	INTERCONNECTIONREDUNDANCY_REDUNDANT InterconnectionRedundancy = "redundant"
)

// All allowed values of InterconnectionRedundancy enum
var AllowedInterconnectionRedundancyEnumValues = []InterconnectionRedundancy{
	"primary",
	"redundant",
}

func (v *InterconnectionRedundancy) UnmarshalJSON(src []byte) error {
	var value string
	err := json.Unmarshal(src, &value)
	if err != nil {
		return err
	}
	enumTypeValue := InterconnectionRedundancy(value)
	for _, existing := range AllowedInterconnectionRedundancyEnumValues {
		if existing == enumTypeValue {
			*v = enumTypeValue
			return nil
		}
	}

	return fmt.Errorf("%+v is not a valid InterconnectionRedundancy", value)
}

// NewInterconnectionRedundancyFromValue returns a pointer to a valid InterconnectionRedundancy
// for the value passed as argument, or an error if the value passed is not allowed by the enum
func NewInterconnectionRedundancyFromValue(v string) (*InterconnectionRedundancy, error) {
	ev := InterconnectionRedundancy(v)
	if ev.IsValid() {
		return &ev, nil
	} else {
		return nil, fmt.Errorf("invalid value '%v' for InterconnectionRedundancy: valid values are %v", v, AllowedInterconnectionRedundancyEnumValues)
	}
}

// IsValid return true if the value is valid for the enum, false otherwise
func (v InterconnectionRedundancy) IsValid() bool {
	for _, existing := range AllowedInterconnectionRedundancyEnumValues {
		if existing == v {
			return true
		}
	}
	return false
}

// Ptr returns reference to Interconnection_redundancy value
func (v InterconnectionRedundancy) Ptr() *InterconnectionRedundancy {
	return &v
}

type NullableInterconnectionRedundancy struct {
	value *InterconnectionRedundancy
	isSet bool
}

func (v NullableInterconnectionRedundancy) Get() *InterconnectionRedundancy {
	return v.value
}

func (v *NullableInterconnectionRedundancy) Set(val *InterconnectionRedundancy) {
	v.value = val
	v.isSet = true
}

func (v NullableInterconnectionRedundancy) IsSet() bool {
	return v.isSet
}

func (v *NullableInterconnectionRedundancy) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableInterconnectionRedundancy(val *InterconnectionRedundancy) *NullableInterconnectionRedundancy {
	return &NullableInterconnectionRedundancy{value: val, isSet: true}
}

func (v NullableInterconnectionRedundancy) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableInterconnectionRedundancy) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
