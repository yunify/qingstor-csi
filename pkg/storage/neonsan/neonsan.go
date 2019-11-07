package neonsan

import (
	"github.com/yunify/qingstor-csi/pkg/storage"
	"github.com/yunify/qingstor-csi/pkg/storage/neonsan/api"
)

func volumeOfNeonsan(v *api.Volume) *storage.Volume {
	return &storage.Volume{
		Status:     &v.Status,
		Size:       &v.Size,
		VolumeName: &v.Name,
		VolumeID:   &v.Name,
	}
}

type neonsan struct {
	confFile string
	poolName string
}

func New(confFile, poolName string) (storage.Provider, error) {
	return &neonsan{
		confFile: confFile,
		poolName: poolName,
	}, nil
}
