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
