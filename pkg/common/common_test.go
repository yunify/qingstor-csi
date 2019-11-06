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
