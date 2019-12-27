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
package service

import (
	"github.com/smartystreets/goconvey/convey"
	"strconv"
	"testing"
)

func TestNewStorageClass(t *testing.T) {
	mapsOK1 := []map[string]string{
		{}, {StorageClassFsTypeName: "ext4",}, {StorageClassReplicaName: "1",},
		{"xxx": "xxx"},
	}
	for _, m := range mapsOK1 {
		sc, err := NewStorageClass(m)
		convey.Convey("new storage class success(1) ", t, func() {
			convey.So(sc, convey.ShouldNotBeNil)
			convey.So(sc.FsType, convey.ShouldNotBeNil)
			convey.So(sc.Replica, convey.ShouldBeGreaterThan,0)
			convey.So(err, convey.ShouldBeNil)
		})
	}

	mapsOK2 := []map[string]string{
		{StorageClassFsTypeName: "ext4", StorageClassReplicaName: "2"},
	}
	for _, m := range mapsOK2 {
		sc, err := NewStorageClass(m)
		convey.Convey("new storage class success(2)", t, func() {
			convey.So(sc, convey.ShouldNotBeNil)
			convey.So(sc.FsType, convey.ShouldEqual,m[StorageClassFsTypeName])
			convey.So(strconv.Itoa(sc.Replica), convey.ShouldEqual,m[StorageClassReplicaName])
			convey.So(err, convey.ShouldBeNil)
		})
	}

	mapsFail := []map[string]string{
		{StorageClassFsTypeName: "xxx"}, {StorageClassReplicaName: "yyy"},
	}
	for _, m := range mapsFail{
		sc, err := NewStorageClass(m)
		convey.Convey("new storage class fail", t, func() {
			convey.So(sc, convey.ShouldBeNil)
			convey.So(err, convey.ShouldNotBeNil)
		})
	}
}

