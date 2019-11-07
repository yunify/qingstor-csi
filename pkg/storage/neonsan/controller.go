package neonsan

import (
	"errors"
	"github.com/yunify/qingstor-csi/pkg/storage"
	"github.com/yunify/qingstor-csi/pkg/storage/neonsan/api"
)

var (
	errorNotImplement = errors.New("method not implement")
	errorNotToCalled  = errors.New("method should not to be called")
)

//requestSize G
func (v *neonsan) CreateVolume(volName string, requestSize int, replicas int) (string, error) {
	_, err := api.CreateVolume(v.confFile, v.poolName, volName, requestSize, replicas)
	if err != nil {
		return "", err
	}

	return volName, nil
}

func (v *neonsan) DeleteVolume(volId string) (err error) {
	_, err = api.DeleteVolume(v.confFile, v.poolName, volId)
	return err
}

func (v *neonsan) FindVolume(volId string) (volInfo *storage.Volume, err error) {
	return v.FindVolumeByName(volId)
}

func (v *neonsan) FindVolumeByName(volName string) (volInfo *storage.Volume, err error) {
	vol, err := api.ListVolume(v.confFile, v.poolName, volName)
	if err != nil {
		return nil, err
	}
	if vol == nil {
		return nil, nil
	}
	return volumeOfNeonsan(vol), nil
}

func (*neonsan) AttachVolume(volId string, instanceId string) (err error) {
	return errorNotToCalled
}

func (*neonsan) DetachVolume(volId string, instanceId string) (err error) {
	return errorNotToCalled
}

func (*neonsan) ResizeVolume(volId string, requestSize int) (err error) {
	return errorNotImplement
}

func (v *neonsan) CloneVolume(volName string, volType int, srcVolId string, zone string) (volId string, err error) {
	return "", errorNotImplement
}
