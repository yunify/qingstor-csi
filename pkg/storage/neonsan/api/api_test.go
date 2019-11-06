package api

import (
	"testing"
)

const (
	configFile = "C:\\Users\\zhangmin\\qbd.conf"
)

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
