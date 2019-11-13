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

package api

import (
	"bou.ke/monkey"
	"bytes"
	"encoding/json"
	"errors"
	"github.com/smartystreets/goconvey/convey"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"testing"
)

var (
	configFile = filepath.Join(os.TempDir(), "qbd.conf")
	poolName   = "ks"

	errMock = errors.New(" error of mock")
)

func TestHttpGet(t *testing.T) {
	request := CreateVolumeRequest{
		Op:       "list_volume",
		PoolName: "",
		Name:     "",
	}
	response := &CreateVolumeResponse{}
	var err error
	guardGetApiUrlFail := monkey.Patch(getApiUrl, func(string) (string, error) { return "", errMock })
	defer guardGetApiUrlFail.Unpatch()
	err = httpGet(configFile, request, response)
	convey.Convey("if get api url error, fail", t, func() {
		convey.So(err, convey.ShouldEqual, errMock)
	})
	guardGetApiUrlFail.Unpatch()

	guardGetApiUrlOK := monkey.Patch(getApiUrl, func(string) (string, error) { return "neonsan-api.com", nil })
	defer guardGetApiUrlOK.Unpatch()

	guardHttpGetFail := monkey.Patch(http.Get, func(string) (*http.Response, error) { return nil, errMock })
	defer guardHttpGetFail.Unpatch()
	err = httpGet(configFile, request, response)
	convey.Convey("if http.Get error, fail", t, func() {
		convey.So(err, convey.ShouldEqual, errMock)
	})
	guardHttpGetFail.Unpatch()

	guardHttpGet404 := monkey.Patch(http.Get, func(string) (*http.Response, error) {
		return &http.Response{StatusCode: 404}, nil
	})
	defer guardHttpGet404.Unpatch()
	err = httpGet(configFile, request, response)
	convey.Convey("if http.Get 404, fail", t, func() {
		convey.So(err.Error(), convey.ShouldNotBeNil)
		convey.So(err.Error(), convey.ShouldContainSubstring, "404")
	})
	guardHttpGet404.Unpatch()

	mockResponse := &CreateVolumeResponse{
		ResponseHeader: ResponseHeader{
			Op:      "list_reply",
			RetCode: 0,
			Reason:  "beautiful",
		}, Id: 1, Size: 10,
	}
	s, _ := json.Marshal(mockResponse)

	httpBody := s
	guardHttpGetOK := monkey.Patch(http.Get, func(string) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200,
			Body:       ioutil.NopCloser(bytes.NewReader(httpBody)),
		}, nil
	})
	defer guardHttpGetOK.Unpatch()
	err = httpGet(configFile, request, response)
	convey.Convey("if neonsan API ok, success", t, func() {
		convey.So(err, convey.ShouldBeNil)
	})

	httpBody = []byte("hello neonsan")
	err = httpGet(configFile, request, response)
	convey.Convey("if http body not json, fail", t, func() {
		convey.So(err, convey.ShouldNotBeNil)
	})

	mockResponse.RetCode = 100
	mockResponse.Reason = "error 100"
	httpBody, _ = json.Marshal(mockResponse)
	guardHttpGetRetCodeError := monkey.Patch(http.Get, func(string) (*http.Response, error) {
		return &http.Response{
			Status:     "",
			StatusCode: 200,
			Body:       ioutil.NopCloser(bytes.NewReader(httpBody)),
		}, nil
	})
	defer guardHttpGetRetCodeError.Unpatch()

	err = httpGet(configFile, request, response)
	convey.Convey("if neonsan api retCode != 0, fail", t, func() {
		convey.So(err.Error(), convey.ShouldEqual, mockResponse.Reason)
	})

}

func TestListVolume(t *testing.T) {

	guardGetApiUrlOK := monkey.Patch(getApiUrl, func(string) (string, error) { return "neonsan-api.com", nil })
	defer guardGetApiUrlOK.Unpatch()

	mockResponse := &ListVolumeResponse{
		ResponseHeader: ResponseHeader{
			Op:      "list_reply",
			RetCode: 0,
			Reason:  "",
		},
		Volumes: []Volume{
			{
				Id: 1,
			},
		},
	}
	httpBody, _ := json.Marshal(mockResponse)

	guardHttpGetOK := monkey.Patch(http.Get, func(string) (*http.Response, error) {
		return &http.Response{
			Status:     "",
			StatusCode: 200,
			Body:       ioutil.NopCloser(bytes.NewReader(httpBody)),
		}, nil
	})
	defer guardHttpGetOK.Unpatch()

	vol, err := ListVolume(configFile, poolName, "yy")
	convey.Convey("list volume success", t, func() {
		convey.So(err, convey.ShouldBeNil)
		convey.So(vol, convey.ShouldNotBeNil)
	})

	mockResponse.Volumes = nil
	httpBody, _ = json.Marshal(mockResponse)
	vol, err = ListVolume(configFile, poolName, "yy")
	convey.Convey("list volume success but volume not exist", t, func() {
		convey.So(err, convey.ShouldBeNil)
		convey.So(vol, convey.ShouldBeNil)
	})

	mockResponse.RetCode = 100
	httpBody, _ = json.Marshal(mockResponse)
	_, err = ListVolume(configFile, poolName, "xx")
	convey.Convey("list volume fail", t, func() {
		convey.So(err, convey.ShouldNotBeNil)
	})

}

func TestCreateVolume(t *testing.T) {
	guardGetApiUrlOK := monkey.Patch(getApiUrl, func(string) (string, error) { return "neonsan-api.com", nil })
	defer guardGetApiUrlOK.Unpatch()

	mockResponse := &CreateVolumeResponse{
		ResponseHeader: ResponseHeader{
			Op:      "list_reply",
			RetCode: 0,
			Reason:  "",
		}, Id: 1, Size: 10,
	}
	httpBody, _ := json.Marshal(mockResponse)

	guardHttpGetOK := monkey.Patch(http.Get, func(string) (*http.Response, error) {
		return &http.Response{
			Status:     "",
			StatusCode: 200,
			Body:       ioutil.NopCloser(bytes.NewReader(httpBody)),
		}, nil
	})
	defer guardHttpGetOK.Unpatch()

	n, err := CreateVolume(configFile, poolName, "xx", 1, 1)
	convey.Convey("create volume success", t, func() {
		convey.So(err, convey.ShouldBeNil)
		convey.So(n, convey.ShouldEqual, mockResponse.Id)
	})

	mockResponse.RetCode = 100
	httpBody, _ = json.Marshal(mockResponse)
	n, err = CreateVolume(configFile, poolName, "xx", 1, 1)
	convey.Convey("create volume fail", t, func() {
		convey.So(err, convey.ShouldNotBeNil)
	})

}

func TestDeleteVolume(t *testing.T) {

	guardGetApiUrlOK := monkey.Patch(getApiUrl, func(string) (string, error) { return "neonsan-api.com", nil })
	defer guardGetApiUrlOK.Unpatch()

	mockResponse := &DeleteVolumeResponse{
		ResponseHeader: ResponseHeader{
			Op:      "list_reply",
			RetCode: 0,
			Reason:  "",
		},
	}
	httpBody, _ := json.Marshal(mockResponse)

	guardHttpGetOK := monkey.Patch(http.Get, func(string) (*http.Response, error) {
		return &http.Response{
			Status:     "",
			StatusCode: 200,
			Body:       ioutil.NopCloser(bytes.NewReader(httpBody)),
		}, nil
	})
	defer guardHttpGetOK.Unpatch()

	n, err := DeleteVolume(configFile, poolName, "xx")
	convey.Convey("delete volume success", t, func() {
		convey.So(err, convey.ShouldBeNil)
		convey.So(n, convey.ShouldEqual, mockResponse.Id)
	})

	mockResponse.RetCode = 100
	httpBody, _ = json.Marshal(mockResponse)
	n, err = DeleteVolume(configFile, poolName, "xx")
	convey.Convey("delete volume fail", t, func() {
		convey.So(err, convey.ShouldNotBeNil)
	})

}

/*
func TestCreateVolume(t *testing.T) {
	volId, err := CreateVolume(configFile, "kube", "happy", 1<<30*10, 1)
	if err != nil {
		t.Error(err)
	}
	t.Log(volId)

}

func TestGetApiUrl(t *testing.T) {
	url, err := getApiUrl(configFile)
	t.Log(url, err)
}

*/
