package mock

import (
	"errors"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/yunify/qingstor-csi/pkg/common"
	"time"
)

var (
	errorNotImplement = errors.New("method not implement")
	errorNotToCalled  = errors.New("method should not to be called")
)

//requestSize G
func (v *mockStorageProvider) CreateVolume(volName string, requestSize int64, replicas int) (volId string, err error) {
	vol, err := v.FindVolumeByName(volName)
	if err != nil{
		return "",err
	}
	if vol != nil{
		return "", errors.New("volume exist")
	}

	volId = common.GenerateHashInEightBytes(time.Now().UTC().String())
	vol = &csi.Volume{
		CapacityBytes: requestSize,
		VolumeId:volId,
	}
	v.volumes[volName] = vol
	return volId,nil
}

func (v *mockStorageProvider) DeleteVolume(volId string) (err error) {
	vol,err := v.FindVolume(volId)
	if vol == nil{
		return errors.New("delete not exist volume")
	}
	delete(v.volumes, volId)
	return nil
}

func (v *mockStorageProvider) FindVolume(volId string) (*csi.Volume, error) {
	for _,vol := range v.volumes {
		if vol.VolumeId == volId{
			return vol,nil
		}
	}
	return nil,nil
}

func (v *mockStorageProvider) FindVolumeByName(volName string) (*csi.Volume, error) {
	return v.volumes[volName],nil
}

func (*mockStorageProvider) AttachVolume(volId string, instanceId string) (err error) {
	return errorNotToCalled
}

func (*mockStorageProvider) DetachVolume(volId string, instanceId string) (err error) {
	return errorNotToCalled
}

func (*mockStorageProvider) ResizeVolume(volId string, requestSize int) (err error) {
	return errorNotImplement
}

func (v *mockStorageProvider) CloneVolume(volName string, volType int, srcVolId string, zone string) (volId string, err error) {
	return "", errorNotImplement
}
