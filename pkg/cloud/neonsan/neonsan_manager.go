package neonsan

import (
	"errors"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/yunify/qingcloud-csi/pkg/cloud"
	"github.com/yunify/qingcloud-sdk-go/service"
)


const (
	CmdQbd               string = "qbd"
	CmdNeonsan           string = "neonsan"
	Pool                 string = "kube"
	//VolumeStatusOk       string = "OK"
	//VolumeStatusError    string = "ERROR"
	//VolumeStatusDegraded string = "DEGRADED"
)

var (
	Pools = []string{"kube"}
	errorNotImplement = errors.New("method not implement")
)

func NewNeonsanVolume(v *Volume) *cloud.Volume {
	volType, zoneID:= 200, "unknown"
	return &cloud.Volume{
		VolumeType: &volType, //mock
		Status:&v.Status,
		Size:&v.Size,
		Instance: &service.Instance{}, //mock
		ZoneID: &zoneID, //mock
		VolumeName:&v.Name,
		VolumeID:&v.Name,
	}
}

type neonsanManager struct {
	volumes   map[string]*cloud.Volume
}


func NewManager() (*neonsanManager, error)  {
	return &neonsanManager{
		volumes: make(map[string]*cloud.Volume),
	},nil
}

func (q *neonsanManager) FindVolume(volId string) (volInfo *cloud.Volume, err error) {
	return q.FindVolumeByName(volId)
}

func (q *neonsanManager) FindVolumeByName(volName string) (volInfo *cloud.Volume, err error) {
	vol, ok :=q.volumes[volName]
	if !ok{
		return nil,nil
	}
	return vol, nil
}

func (q *neonsanManager) CreateVolume(volName string, requestSize int, replicas int, volType int, zone string) (string,error) {
	nVol := &Volume{
		Name:volName,
		Size:requestSize,
		ReplicationCount:requestSize,
	}
	q.volumes[volName] = NewNeonsanVolume(nVol)
	return volName,nil
}

func (q *neonsanManager) DeleteVolume(volId string) (err error) {
	delete(q.volumes, volId)
	return nil
}

func (*neonsanManager) FindSnapshot(snapId string) (snapInfo *service.Snapshot, err error) {
	return nil, errorNotImplement
}

func (*neonsanManager) FindSnapshotByName(snapName string) (snapInfo *service.Snapshot, err error) {
	return nil, errorNotImplement
}

func (*neonsanManager) CreateSnapshot(snapName string, volId string) (snapId string, err error) {
	panic("implement me")
}

func (*neonsanManager) DeleteSnapshot(snapId string) (err error) {
	panic("implement me")
}

func (*neonsanManager) CreateVolumeFromSnapshot(volName string, snapId string, zone string) (volId string, err error) {
	panic("implement me")
}

func (*neonsanManager) AttachVolume(volId string, instanceId string) (err error) {
	return nil
}

func (*neonsanManager) DetachVolume(volId string, instanceId string) (err error) {
	return nil
}

func (q *neonsanManager) NodeAttachVolume(volId string) error {
	vol, ok := q.volumes[volId]
	if !ok {
		return errors.New("vol not exist")
	}
	device := "/tmp/xx"
	vol.Device = &device
	return nil
}

func (q *neonsanManager) NodeDetachVolume(volId string) error {
	return nil
}

func (q *neonsanManager) NodeGetDevice(volId string) (string, error) {
	return "/tmp/xxx",nil
}

func (*neonsanManager) ResizeVolume(volId string, requestSize int) (err error) {
	return errorNotImplement
}

func (*neonsanManager) CloneVolume(volName string, volType int, srcVolId string, zone string) (volId string, err error) {
	return "", errorNotImplement
}

func (*neonsanManager) FindInstance(instanceId string) (instanceInfo *service.Instance, err error) {
	return nil, errorNotImplement
}

func (*neonsanManager) GetZone() (zoneName string) {
	return ""
}

func (*neonsanManager) GetZoneList() (zoneNameList []string, err error) {
	return []string{"unknown"} ,nil
}

func (*neonsanManager) Probe() error {
	return nil
}
func (q *neonsanManager) GetTopology(instanceId string) (*csi.Topology, error) {
	return &csi.Topology{},nil
}

func (*neonsanManager) FindTag(tagId string) (tagInfo *service.Tag, err error) {
	return &service.Tag{TagID:&tagId}, nil
}

func (*neonsanManager) IsValidTags(tagsId []string) bool {
	return true
}

func (*neonsanManager) AttachTags(tagsId []string, resourceId string, resourceType string) (err error) {
	return nil
}

