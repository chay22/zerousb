// Copyright 2013 Google Inc.  All rights reserved.
// Copyright 2016 the gousb Authors.  All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package zerousb

import "strconv"

// Class represents a USB-IF (Implementers Forum) class or subclass code.
type Class uint8

// Standard classes defined by USB spec, see https://www.usb.org/defined-class-codes
const (
	ClassPerInterface       Class = 0x00
	ClassAudio              Class = 0x01
	ClassComm               Class = 0x02
	ClassHID                Class = 0x03
	ClassPhysical           Class = 0x05
	ClassImage              Class = 0x06
	ClassPTP                Class = ClassImage // legacy name for image
	ClassPrinter            Class = 0x07
	ClassMassStorage        Class = 0x08
	ClassHub                Class = 0x09
	ClassData               Class = 0x0a
	ClassSmartCard          Class = 0x0b
	ClassContentSecurity    Class = 0x0d
	ClassVideo              Class = 0x0e
	ClassPersonalHealthcare Class = 0x0f
	ClassAudioVideo         Class = 0x10
	ClassBillboard          Class = 0x11
	ClassUSBTypeCBridge     Class = 0x12
	ClassDiagnosticDevice   Class = 0xdc
	ClassWireless           Class = 0xe0
	ClassMiscellaneous      Class = 0xef
	ClassApplication        Class = 0xfe
	ClassVendorSpec         Class = 0xff
)

var classDescription = map[Class]string{
	ClassPerInterface:       "per-interface",
	ClassAudio:              "audio",
	ClassComm:               "communications",
	ClassHID:                "human interface device",
	ClassPhysical:           "physical",
	ClassImage:              "image",
	ClassPrinter:            "printer",
	ClassMassStorage:        "mass storage",
	ClassHub:                "hub",
	ClassData:               "data",
	ClassSmartCard:          "smart card",
	ClassContentSecurity:    "content security",
	ClassVideo:              "video",
	ClassPersonalHealthcare: "personal healthcare",
	ClassAudioVideo:         "audio/video",
	ClassBillboard:          "billboard",
	ClassUSBTypeCBridge:     "USB type-C bridge",
	ClassDiagnosticDevice:   "diagnostic device",
	ClassWireless:           "wireless",
	ClassMiscellaneous:      "miscellaneous",
	ClassApplication:        "application-specific",
	ClassVendorSpec:         "vendor-specific",
}

func (c Class) String() string {
	if d, ok := classDescription[c]; ok {
		return d
	}
	return strconv.Itoa(int(c))
}

// Protocol is the interface class protocol, qualified by the values
// of interface class and subclass.
type Protocol uint8

func (p Protocol) String() string {
	return strconv.Itoa(int(p))
}

// DescriptorType identifies the type of a USB descriptor.
type DescriptorType uint8

// Descriptor types defined by the USB spec.
const (
	DescriptorTypeDevice    DescriptorType = 0x1
	DescriptorTypeConfig    DescriptorType = 0x2
	DescriptorTypeString    DescriptorType = 0x3
	DescriptorTypeInterface DescriptorType = 0x4
	DescriptorTypeEndpoint  DescriptorType = 0x5
	DescriptorTypeHID       DescriptorType = 0x21
	DescriptorTypeReport    DescriptorType = 0x22
	DescriptorTypePhysical  DescriptorType = 0x23
	DescriptorTypeHub       DescriptorType = 0x29
)

var descriptorTypeDescription = map[DescriptorType]string{
	DescriptorTypeDevice:    "device",
	DescriptorTypeConfig:    "configuration",
	DescriptorTypeString:    "string",
	DescriptorTypeInterface: "interface",
	DescriptorTypeEndpoint:  "endpoint",
	DescriptorTypeHID:       "HID",
	DescriptorTypeReport:    "HID report",
	DescriptorTypePhysical:  "physical",
	DescriptorTypeHub:       "hub",
}

func (dt DescriptorType) String() string {
	return descriptorTypeDescription[dt]
}

// EndpointDirection defines the direction of data flow - IN (device to host)
// or OUT (host to device).
type EndpointDirection bool

const (
	endpointNumMask       = 0x0f
	endpointDirectionMask = 0x80
	// EndpointDirectionIn marks data flowing from device to host.
	EndpointDirectionIn EndpointDirection = true
	// EndpointDirectionOut marks data flowing from host to device.
	EndpointDirectionOut EndpointDirection = false
)

var endpointDirectionDescription = map[EndpointDirection]string{
	EndpointDirectionIn:  "IN",
	EndpointDirectionOut: "OUT",
}

func (ed EndpointDirection) String() string {
	return endpointDirectionDescription[ed]
}

// TransferType defines the endpoint transfer type.
type TransferType uint8

// Transfer types defined by the USB spec.
const (
	TransferTypeControl     TransferType = 0x0
	TransferTypeIsochronous TransferType = 0x1
	TransferTypeBulk        TransferType = 0x2
	TransferTypeInterrupt   TransferType = 0x3
	transferTypeMask                     = 0x03
)

var transferTypeDescription = map[TransferType]string{
	TransferTypeControl:     "control",
	TransferTypeIsochronous: "isochronous",
	TransferTypeBulk:        "bulk",
	TransferTypeInterrupt:   "interrupt",
}

// String returns a human-readable name of the endpoint transfer type.
func (tt TransferType) String() string {
	return transferTypeDescription[tt]
}

// IsoSyncType defines the isochronous transfer synchronization type.
type IsoSyncType uint8

// Synchronization types defined by the USB spec.
const (
	IsoSyncTypeNone     IsoSyncType = 0x0 << 2
	IsoSyncTypeAsync    IsoSyncType = 0x1 << 2
	IsoSyncTypeAdaptive IsoSyncType = 0x2 << 2
	IsoSyncTypeSync     IsoSyncType = 0x3 << 2
	isoSyncTypeMask                 = 0x0C
)

var isoSyncTypeDescription = map[IsoSyncType]string{
	IsoSyncTypeNone:     "unsynchronized",
	IsoSyncTypeAsync:    "asynchronous",
	IsoSyncTypeAdaptive: "adaptive",
	IsoSyncTypeSync:     "synchronous",
}

// String returns a human-readable description of the synchronization type.
func (ist IsoSyncType) String() string {
	return isoSyncTypeDescription[ist]
}

// UsageType defines the transfer usage type for isochronous and interrupt
// transfers.
type UsageType uint8

// Usage types for iso and interrupt transfers, defined by the USB spec.
const (
	// Note: USB3.0 defines usage type for both isochronous and interrupt
	// endpoints, with the same constants representing different usage types.
	// UsageType constants do not correspond to bmAttribute values.
	UsageTypeUndefined UsageType = iota
	IsoUsageTypeData
	IsoUsageTypeFeedback
	IsoUsageTypeImplicit
	InterruptUsageTypePeriodic
	InterruptUsageTypeNotification
	usageTypeMask = 0x30
)

var usageTypeDescription = map[UsageType]string{
	UsageTypeUndefined:             "undefined usage",
	IsoUsageTypeData:               "data",
	IsoUsageTypeFeedback:           "feedback",
	IsoUsageTypeImplicit:           "implicit data",
	InterruptUsageTypePeriodic:     "periodic",
	InterruptUsageTypeNotification: "notification",
}

func (ut UsageType) String() string {
	return usageTypeDescription[ut]
}

// Control request type bit fields as defined in the USB spec. All values are
// of uint8 type.  These constants can be used with Device.Control() method to
// specify the type and destination of the control request, e.g.
// `dev.Control(ControlOut|ControlVendor|ControlDevice, ...)`.
const (
	ControlIn  = 0x80
	ControlOut = 0x0

	// "Standard" is explicitly omitted, as functionality of standard requests
	// is exposed through higher level operations of gousb.
	ControlClass  = 0x20
	ControlVendor = 0x40
	// "Reserved" is explicitly omitted, should not be used.

	ControlDevice    = 0x0
	ControlInterface = 0x1
	ControlEndpoint  = 0x2
	ControlOther     = 0x3
)

// Speed identifies the speed of the device.
type Speed int

// Device speeds as defined in the USB spec.
const (
	SpeedUnknown Speed = 0x0
	SpeedLow     Speed = 0x1
	SpeedFull    Speed = 0x2
	SpeedHigh    Speed = 0x3
	SpeedSuper   Speed = 0x4
)

var deviceSpeedDescription = map[Speed]string{
	SpeedUnknown: "unknown",
	SpeedLow:     "low",
	SpeedFull:    "full",
	SpeedHigh:    "high",
	SpeedSuper:   "super",
}

// String returns a human-readable name of the device speed.
func (s Speed) String() string {
	return deviceSpeedDescription[s]
}

const (
	selfPoweredMask  = 0x40
	remoteWakeupMask = 0x20
)

// Milliamperes is a unit of electric current consumption.
type Milliamperes uint
