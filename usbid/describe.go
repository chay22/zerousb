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

// Package usbid provides human-readable text output for the usb package.
//
// On load, the usbid package parses an embedded mapping of vendors/products
// and class/subclass/protocols.  They can also be loaded from a URL or from
// a reader.
//
// The bread and butter of this package are the following two functions:
//   Describe - Pretty-print the vendor and product of a device descriptor
//   Classify - Pretty-print the class/protocol info for a device/interface
package usbid

import (
	"fmt"

	"github.com/chay22/zerousb"
)

// Describe returns a human readable string describing the vendor and product
// of the given device.
//
// The given val must be one of the following:
//   - zerousb.DeviceInfo       "Product (Vendor)"
func Describe(val interface{}) string {
	switch val := val.(type) {
	case zerousb.DeviceInfo:
		if v, ok := Vendors[val.VendorID]; ok {
			if d, ok := v.Product[val.ProductID]; ok {
				return fmt.Sprintf("%s (%s)", v, d)
			}
			return fmt.Sprintf("%s - Unknown", v)
		}
		return fmt.Sprintf("Unknown %d:%d", val.VendorID, val.ProductID)
	}
	return fmt.Sprintf("Unknown (%T)", val)
}

// Classify returns a human-readable string describing the class, subclass,
// and protocol associated with a device or interface.
//
// The given val must be one of the following:
//   - zerousb.DeviceInfo       "Class (SubClass) Protocol"
func Classify(val interface{}) string {
	var (
		class, sub uint8
		proto      uint8
	)
	switch val := val.(type) {
	case zerousb.DeviceInfo:
		class, sub, proto = val.Class, val.SubClass, val.Protocol
	default:
		return fmt.Sprintf("Unknown (%T)", val)
	}

	if c, ok := Classes[class]; ok {
		if s, ok := c.SubClass[sub]; ok {
			if p, ok := s.Protocol[proto]; ok {
				return fmt.Sprintf("%s (%s) %s", c, s, p)
			}
			return fmt.Sprintf("%s (%s)", c, s)
		}
		return fmt.Sprintf("%s", c)
	}
	return fmt.Sprintf("Unknown %d.%d.%d", class, sub, proto)
}
