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

package usbid

import (
	"log"
	"strings"
)

var (
	// Vendors stores the vendor and product ID mappings.
	Vendors map[uint16]*Vendor

	// Classes stores the class, subclass and protocol mappings.
	Classes map[uint8]*Class
)

//go:generate go run regen/regen.go --template regen/load_data.go.tpl -o load_data.go

func init() {
	ids, cls, err := ParseIDs(strings.NewReader(usbIDListData))
	if err != nil {
		log.Printf("usbid: failed to parse: %s", err)
		return
	}

	Vendors = ids
	Classes = cls
}
