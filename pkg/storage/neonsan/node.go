package neonsan

import "github.com/yunify/qingstor-csi/pkg/storage/neonsan/qbd"

func (v *neonsan) NodeAttachVolume(volId string) error {
	return qbd.AttachVolume(v.confFile, v.poolName, volId)
}

func (v *neonsan) NodeDetachVolume(volId string) error {
	return qbd.DetachVolume(v.confFile, v.poolName, volId)
}

func (v *neonsan) NodeGetDevice(volId string) (string, error) {
	attachInfo, err := qbd.FindAttachedVolumeWithoutPool(volId)
	if err != nil {
		return "", err
	}
	if attachInfo != nil {
		return attachInfo.Device, nil
	}
	return "", nil
}
