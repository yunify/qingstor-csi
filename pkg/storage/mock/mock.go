package mock

import (
	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/yunify/qingstor-csi/pkg/storage"
)

type attachVolume struct {
	vol    *csi.Volume
	device string
}

type mockStorageProvider struct {
	volumes         map[string]*csi.Volume
	attachedVolumes map[string]*attachVolume
}

func New() storage.Provider  {
	return &mockStorageProvider{
		volumes:         make(map[string]*csi.Volume),
		attachedVolumes: make(map[string]*attachVolume),
	}
}
