package mock

import (
	"errors"
	"github.com/yunify/qingstor-csi/pkg/common"
	"time"
)

//var deviceNo = 50

func (v *mockStorageProvider) NodeAttachVolume(volId string) error {
	_, ok := v.attachedVolumes[volId]
	if ok {
		return errors.New("volume already attached")
	}
	vol, err := v.FindVolume(volId)
	if err != nil{
		return err
	}
	//deviceNo ++
	v.attachedVolumes[volId] = &attachVolume{
		vol:vol,
		device: common.GenerateHashInEightBytes(time.Now().UTC().String()),
		//device: strconv.Itoa(deviceNo),
	}
	return nil
}

func (v *mockStorageProvider) NodeDetachVolume(volId string) error {
	_, ok := v.attachedVolumes[volId]
	if !ok {
		return errors.New("volume not attached")
	}
	delete(v.attachedVolumes,volId)
	return nil
}

func (v *mockStorageProvider) NodeGetDevice(volId string) (string, error) {
	attachVol, ok := v.attachedVolumes[volId]
	if ok && attachVol != nil{
		return  attachVol.device,nil
	}
	return "" , errors.New("vol not found")
}
