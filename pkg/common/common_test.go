/*
Copyright (C) 2018 Yunify, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this work except in compliance with the License.
You may obtain a copy of the License in the LICENSE file, or at:

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package common

import (
	"github.com/container-storage-interface/spec/lib/go/csi"
	"testing"
)

func TestGenerateHashInEightBytes(t *testing.T) {
	tests := []struct {
		name  string
		input string
		hash  string
	}{
		{
			name:  "normal",
			input: "snapshot",
			hash:  "2aa38b8d",
		},
		{
			name:  "empty input",
			input: "",
			hash:  "811c9dc5",
		},
	}
	for _, v := range tests {
		res := GenerateHashInEightBytes(v.input)
		if v.hash != res {
			t.Errorf("name %s: expect %s but actually %s", v.name, v.hash, res)
		}
	}
}

func TestIsValidCapacityBytes(t *testing.T) {
	tests := []struct {
		name     string
		bytes    int64
		capRange *csi.CapacityRange
		isValid  bool
	}{
		{
			name:  "normal",
			bytes: 10 * Gib,
			capRange: &csi.CapacityRange{
				RequiredBytes: 10 * Gib,
				LimitBytes:    10 * Gib,
			},
			isValid: true,
		},
		{
			name:  "invalid range",
			bytes: 10 * Gib,
			capRange: &csi.CapacityRange{
				RequiredBytes: 11 * Gib,
				LimitBytes:    10 * Gib,
			},
			isValid: false,
		},
		{
			name:     "empty range",
			bytes:    10 * Gib,
			capRange: &csi.CapacityRange{},
			isValid:  true,
		},
		{
			name:     "nil range",
			bytes:    10 * Gib,
			capRange: nil,
			isValid:  true,
		},
		{
			name:  "without floor",
			bytes: 10 * Gib,
			capRange: &csi.CapacityRange{
				LimitBytes: 10*Gib + 1,
			},
			isValid: true,
		},
		{
			name:  "invalid floor",
			bytes: 11 * Gib,
			capRange: &csi.CapacityRange{
				RequiredBytes: 11*Gib + 1,
			},
			isValid: false,
		},
		{
			name:  "without ceil",
			bytes: 14 * Gib,
			capRange: &csi.CapacityRange{
				RequiredBytes: 14 * Gib,
			},
			isValid: true,
		},
		{
			name:  "invalid ceil",
			bytes: 14 * Gib,
			capRange: &csi.CapacityRange{
				LimitBytes: 14*Gib - 1,
			},
			isValid: false,
		},
	}
	for _, test := range tests {
		res := IsValidCapacityBytes(test.bytes, test.capRange)
		if test.isValid != res {
			t.Errorf("name %s: expect %t, but actually %t", test.name, test.isValid, res)
		}
	}
}
