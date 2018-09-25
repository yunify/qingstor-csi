/*
Copyright 2018 Yunify, Inc.

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

package util

import (
	"reflect"
	"testing"
)

func TestFormatVolumeSize(t *testing.T) {
	tests := []struct {
		name      string
		inputSize int64
		step      int64
		outSize   int64
	}{
		{
			name:      "format 4Gi, step 1Gi",
			inputSize: 4294967296,
			step:      Gib,
			outSize:   4294967296,
		},
		{
			name:      "format 4Gi, step 10Gi",
			inputSize: 4294967296,
			step:      Gib * 10,
			outSize:   Gib * 10,
		},
		{
			name:      "format 4Gi, step 3Gi",
			inputSize: 4294967296,
			step:      Gib * 3,
			outSize:   Gib * 6,
		},
	}
	for _, v := range tests {
		out := FormatVolumeSize(v.inputSize, v.step)
		if v.outSize != out {
			t.Errorf("name %s: expect %d, but actually %d", v.name, v.outSize, out)
		}
	}
}

func TestParseIntToDec(t *testing.T) {
	tests := []struct {
		name string
		hex  string
		dec  string
	}{
		{
			name: "success parse",
			hex:  "0x3ff7000000",
			dec:  "274726912000",
		},
		{
			name: "failed parse",
			hex:  "321",
			dec:  "321",
		},
	}
	for _, v := range tests {
		ret := ParseIntToDec(v.hex)
		if v.dec != ret {
			t.Errorf("name [%s]: expect [%s], but actually [%s]", v.name, v.dec, ret)
		}

	}
}

func TestGetList(t *testing.T) {
	tests := []struct {
		name string
		str  string
		list []string
	}{
		{
			name: "normal",
			str:  "csi , kube, vol ",
			list: []string{
				"csi",
				"kube",
				"vol",
			},
		},
	}
	for _, v := range tests {
		list := GetList(v.str)
		if !reflect.DeepEqual(v.list, list) {
			t.Errorf("name [%s]: expect [%v], but actually [%v]", v.name, v.list, list)
		}
	}
}
