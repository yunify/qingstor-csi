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
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("Util", func() {
	DescribeTable("format volume size",
		func(inputSize int64, step int64, outSize int64) {
			out := FormatVolumeSize(inputSize, step)
			Expect(out).To(Equal(outSize))
		},
		Entry("format 4Gi, step 1Gi",
			int64(4294967296),
			int64(Gib),
			int64(4294967296)),
		Entry("format 4Gi, step 10Gi",
			int64(4294967296),
			int64(Gib*10),
			int64(Gib*10)),
		Entry("format 4Gi, step 3Gi",
			int64(4294967296),
			int64(Gib*3),
			int64(Gib*6)),
	)

	DescribeTable("is valid file system type",
		func(fs string, isValid bool) {
			out := IsValidFileSystemType(fs)
			Expect(out).To(Equal(isValid))
		},
		Entry("ext3", FileSystemExt3, true),
		Entry("ext4", FileSystemExt4, true),
		Entry("xfs", FileSystemXfs, true),
		Entry("failed", "NTFS", false),
	)

	DescribeTable("parse int to dec",
		func(hex, dec string) {
			out := ParseIntToDec(hex)
			Expect(out).To(Equal(dec))
		},
		Entry("success parse", "0x3ff7000000", "274726912000"),
		Entry("failed parse", "321", "321"),
	)

	DescribeTable("get list",
		func(str string, list []string) {
			out := GetList(str)
			Expect(out).To(Equal(list))
		},
		Entry("three pools", "csi,kube,vol ", []string{"csi", "kube", "vol"}),
		Entry("one pool", "kube", []string{"kube"}),
		Entry("zero pools", "", []string{}),
	)
})
