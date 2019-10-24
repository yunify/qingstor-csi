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

package disk

import (
	"errors"
	"fmt"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/yunify/qingcloud-csi/pkg/cloud"
	qcclient "github.com/yunify/qingcloud-sdk-go/client"
	qcconfig "github.com/yunify/qingcloud-sdk-go/config"
	qcservice "github.com/yunify/qingcloud-sdk-go/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/klog"
)

var _ cloud.CloudManager = &qingCloudManager{}


func NewQingCloudVolume(v *qcservice.Volume) *cloud.Volume {
	return &cloud.Volume{
		VolumeType: v.VolumeType,
		Status:     v.Status,
		Size:       v.Size,
		Instance:   v.Instance,
		ZoneID:     v.ZoneID,

		VolumeName: v.VolumeName,
		VolumeID:   v.VolumeID,
	}
}

type qingCloudManager struct {
	instanceService *qcservice.InstanceService
	snapshotService *qcservice.SnapshotService
	volumeService   *qcservice.VolumeService
	jobService      *qcservice.JobService
	cloudService    *qcservice.QingCloudService
	tagService      *qcservice.TagService
}

func NewQingCloudManagerFromConfig(config *qcconfig.Config) (*qingCloudManager, error) {
	// initial qingcloud iaas service
	qs, err := qcservice.Init(config)
	if err != nil {
		return nil, err
	}
	// create services
	is, _ := qs.Instance(config.Zone)
	ss, _ := qs.Snapshot(config.Zone)
	vs, _ := qs.Volume(config.Zone)
	js, _ := qs.Job(config.Zone)
	ts, _ := qs.Tag(config.Zone)

	// initial cloud manager
	cm := qingCloudManager{
		instanceService: is,
		snapshotService: ss,
		volumeService:   vs,
		jobService:      js,
		cloudService:    qs,
		tagService:      ts,
	}
	klog.Infof("Succeed to initial cloud manager")
	return &cm, nil
}

// NewCloudManagerFromFile
// Create cloud manager from file
func NewQingCloudManagerFromFile(filePath string) (*qingCloudManager, error) {
	// create config
	config, err := cloud.ReadConfigFromFile(filePath)
	if err != nil {
		return nil, err
	}
	return NewQingCloudManagerFromConfig(config)
}

// Find snapshot by snapshot id
// Return: 	nil,	nil: 	not found snapshot
//			snapshot, nil: 	found snapshot
//			nil, 	error:	internal error
func (q *qingCloudManager) FindSnapshot(id string) (snapshot *qcservice.Snapshot, err error) {
	verboseMode := cloud.EnableDescribeSnapshotVerboseMode
	// Set DescribeSnapshot input
	input := &qcservice.DescribeSnapshotsInput{
		Snapshots: []*string{&id},
		Verbose:   &verboseMode,
	}
	// Call describe snapshot
	output, err := q.snapshotService.DescribeSnapshots(input)
	// 1. Error is not equal to nil.
	if err != nil {
		return nil, err
	}
	// 2. Return code is not equal to 0.
	if *output.RetCode != 0 {
		klog.Errorf("Ret code: %d, message: %s", *output.RetCode, *output.Message)
		return nil, fmt.Errorf("call IaaS DescribeSnapshot err: snapshot id %s in %s",
			id, q.snapshotService.Config.Zone)
	}
	switch *output.TotalCount {
	// Not found snapshot
	case 0:
		return nil, nil
	// Found one snapshot
	case 1:
		if *output.SnapshotSet[0].Status == cloud.SnapshotStatusCeased ||
			*output.SnapshotSet[0].Status == cloud.SnapshotStatusDeleted {
			return nil, nil
		}
		return output.SnapshotSet[0], nil
	// Found duplicate snapshots
	default:
		return nil,
			fmt.Errorf("call IaaS DescribeSnapshot err: find duplicate snapshot, snapshot id %s in %s",
				id, q.snapshotService.Config.Zone)
	}
}

// Find snapshot by snapshot name
// In Qingcloud IaaS platform, it is possible that two snapshots have the same name.
// In Kubernetes, the CO will set a unique PV name.
// CSI driver take the PV name as a snapshot name.
// Return: 	nil, 		nil: 	not found snapshots
//			snapshots,	nil:	found snapshot
//			nil,		error:	internal error
func (q *qingCloudManager) FindSnapshotByName(name string) (snapshot *qcservice.Snapshot, err error) {
	if len(name) == 0 {
		return nil, nil
	}
	verboseMode := cloud.EnableDescribeSnapshotVerboseMode
	// Set input arguments
	input := &qcservice.DescribeSnapshotsInput{
		SearchWord: &name,
		Verbose:    &verboseMode,
	}
	// Call DescribeSnapshot
	output, err := q.snapshotService.DescribeSnapshots(input)
	// Handle error
	if err != nil {
		return nil, err
	}
	if *output.RetCode != 0 {
		klog.Errorf("Ret code: %d, message: %s", *output.RetCode, *output.Message)
		return nil, fmt.Errorf("call IaaS DescribeSnapshots err: snapshot name %s in %s",
			name, q.snapshotService.Config.Zone)
	}
	// Not found snapshots
	for _, v := range output.SnapshotSet {
		if *v.SnapshotName == name && *v.Status != cloud.SnapshotStatusCeased && *v.Status != cloud.SnapshotStatusDeleted {
			return v, nil
		}
	}
	return nil, nil
}

// CreateSnapshot
// 1. format snapshot size
// 2. create snapshot
// 3. wait job
func (q *qingCloudManager) CreateSnapshot(snapshotName string, resourceId string) (snapshotId string, err error) {
	// 0. Set CreateSnapshot args
	isFull := int(cloud.SnapshotFull)
	// set input value
	input := &qcservice.CreateSnapshotsInput{
		SnapshotName: &snapshotName,
		IsFull:       &isFull,
		Resources:    []*string{&resourceId},
	}

	// 1. Create snapshot
	klog.Infof("Call IaaS CreateSnapshot request snapshot name: %s, zone: %s, resource id %s, is full snapshot %T",
		*input.SnapshotName, q.GetZone(), *input.Resources[0], *input.IsFull == cloud.SnapshotFull)
	output, err := q.snapshotService.CreateSnapshots(input)
	if err != nil {
		return "", err
	}
	// check output
	if *output.RetCode != 0 {
		klog.Errorf("Ret code: %d, message: %s", *output.RetCode, *output.Message)
		return "", fmt.Errorf("call IaaS CreateSnapshot error: %s", *output.Message)
	}
	snapshotId = *output.Snapshots[0]
	klog.Infof("Call IaaS CreateSnapshots snapshot name %s snapshot id %s succeed", snapshotName, snapshotId)
	return snapshotId, nil
}

// DeleteSnapshot
// 1. delete snapshot by id
// 2. wait job
func (q *qingCloudManager) DeleteSnapshot(snapshotId string) error {
	// set input value
	input := &qcservice.DeleteSnapshotsInput{
		Snapshots: []*string{&snapshotId},
	}
	// delete snapshot
	klog.Infof("Call IaaS DeleteSnapshot request id: %s, zone: %s",
		snapshotId, *q.snapshotService.Properties.Zone)
	output, err := q.snapshotService.DeleteSnapshots(input)
	if err != nil {
		return err
	}
	// wait job
	klog.Infof("Call IaaS WaitJob %s", *output.JobID)
	if err := q.waitJob(*output.JobID); err != nil {
		return err
	}
	// check output
	if *output.RetCode != 0 {
		klog.Errorf("Ret code: %d, message: %s", *output.RetCode, *output.Message)
		return fmt.Errorf(*output.Message)
	}
	klog.Infof("Call IaaS DeleteSnapshot %s succeed", snapshotId)
	return nil
}

// Find volume by volume ID
// Return: 	nil,	nil: 	not found volumes
//			volume, nil: 	found volume
//			nil, 	error:	internal error
func (q *qingCloudManager) FindVolume(id string) (volInfo *cloud.Volume, err error) {
	// Set DescribeVolumes input
	input := &qcservice.DescribeVolumesInput{
		Volumes: []*string{&id},
	}
	// Call describe volume
	output, err := q.volumeService.DescribeVolumes(input)
	// Error:
	// 1. Error is not equal to nil.
	if err != nil {
		return nil, err
	}
	// 2. Return code is not equal to 0.
	if *output.RetCode != 0 {
		klog.Errorf("Ret code: %d, message: %s", *output.RetCode, *output.Message)
		return nil,
			fmt.Errorf("call IaaS DescribeVolumes err: volume id %s in %s", id, q.volumeService.Config.Zone)
	}
	switch *output.TotalCount {
	// Not found volumes
	case 0:
		return nil, nil
	// Found one volume
	case 1:
		if *output.VolumeSet[0].Status == cloud.DiskStatusCeased || *output.VolumeSet[0].
			Status == cloud.DiskStatusDeleted {
			return nil, nil
		}
		return NewQingCloudVolume(output.VolumeSet[0]), nil
	// Found duplicate volumes
	default:
		return nil,
			fmt.Errorf("call IaaS DescribeVolumes err: find duplicate volumes, volume id %s in %s",
				id, q.volumeService.Config.Zone)
	}
}

// Find volume by volume name
// In Qingcloud IaaS platform, it is possible that two volumes have the same name.
// In Kubernetes, the CO will set a unique PV name.
// CSI driver take the PV name as a volume name.
// Return: 	nil, 		nil: 	not found volumes
//			volumes,	nil:	found volume
//			nil,		error:	internal error
func (q *qingCloudManager) FindVolumeByName(name string) (volume *cloud.Volume, err error) {
	if len(name) == 0 {
		return nil, nil
	}
	// Set input arguments
	input := &qcservice.DescribeVolumesInput{
		SearchWord: &name,
	}
	// Call DescribeVolumes
	output, err := q.volumeService.DescribeVolumes(input)
	// Handle error
	if err != nil {
		return nil, err
	}
	if *output.RetCode != 0 {
		klog.Errorf("Ret code: %d, message: %s", *output.RetCode, *output.Message)
		return nil, fmt.Errorf("call IaaS DescribeVolumes err: volume name %s in %s",
			name, q.volumeService.Config.Zone)
	}
	// Not found volumes
	for _, v := range output.VolumeSet {
		if *v.VolumeName != name {
			continue
		}
		if *v.Status == cloud.DiskStatusCeased || *v.Status == cloud.DiskStatusDeleted {
			continue
		}
		return NewQingCloudVolume(v), nil
	}
	return nil, nil
}

// CreateVolume
// 1. format volume size
// 2. create volume
// 3. wait job
func (q *qingCloudManager) CreateVolume(volName string, requestSize int, replicas int, volType int, zone string) (
	newVolId string, err error) {
	// 0. Set CreateVolume args
	// create volume count
	count := 1
	// volume replicas
	replStr := cloud.DiskReplicaTypeName[replicas]
	// set input value
	input := &qcservice.CreateVolumesInput{
		Count:      &count,
		Repl:       &replStr,
		Size:       &requestSize,
		VolumeName: &volName,
		VolumeType: &volType,
		Zone:       &zone,
	}
	// 1. Create volume
	klog.Infof("Call IaaS CreateVolume request name: %s, size: %d GB, type: %d, zone: %s, count: %d, replica: %s",
		*input.VolumeName, *input.Size, *input.VolumeType, *input.Zone, *input.Count, *input.Repl)
	output, err := q.volumeService.CreateVolumes(input)
	if err != nil {
		return "", err
	}
	// wait job
	klog.Infof("Call IaaS WaitJob %s", *output.JobID)
	if err := q.waitJob(*output.JobID); err != nil {
		return "", err
	}
	// check output
	if *output.RetCode != 0 {
		klog.Errorf("Ret code: %d, message: %s", *output.RetCode, *output.Message)
		return "", errors.New(*output.Message)
	}
	newVolId = *output.Volumes[0]
	klog.Infof("Call IaaS CreateVolume name %s id %s succeed", volName, newVolId)
	return newVolId, nil
}

// CreateVolumeFromSnapshot
// In QingCloud, the volume size created from snapshot is equal to original volume.
func (q *qingCloudManager) CreateVolumeFromSnapshot(volName string, snapId string, zone string) (
	volId string, err error) {
	input := &qcservice.CreateVolumeFromSnapshotInput{
		VolumeName: &volName,
		Snapshot:   &snapId,
		Zone:       &zone,
	}
	klog.Infof("Call IaaS CreateVolumeFromSnapshot request volume name: %s, snapshot id: %s",
		*input.VolumeName, *input.Snapshot)
	output, err := q.snapshotService.CreateVolumeFromSnapshot(input)
	if err != nil {
		return "", err
	}
	// wait job
	klog.Infof("Call IaaS WaitJob %s", *output.JobID)
	if err := q.waitJob(*output.JobID); err != nil {
		return "", err
	}
	// check output
	if *output.RetCode != 0 {
		klog.Errorf("Ret code: %d, message: %s", *output.RetCode, *output.Message)
		return "", fmt.Errorf(*output.Message)
	}
	klog.Infof("Call IaaS CreateVolumeFromSnapshot succeed, volume id %s", *output.VolumeID)
	return *output.VolumeID, nil
}

// DeleteVolume
// 1. delete volume by id
// 2. wait job
func (q *qingCloudManager) DeleteVolume(id string) error {
	// set input value
	input := &qcservice.DeleteVolumesInput{
		Volumes: []*string{&id},
	}
	// delete volume
	klog.Infof("Call IaaS DeleteVolume request id: %s, zone: %s",
		id, *q.volumeService.Properties.Zone)
	output, err := q.volumeService.DeleteVolumes(input)
	if err != nil {
		return err
	}
	// wait job
	klog.Infof("Call IaaS WaitJob %s", *output.JobID)
	if err := q.waitJob(*output.JobID); err != nil {
		return err
	}
	// check output
	if *output.RetCode != 0 {
		klog.Errorf("Ret code: %d, message: %s", *output.RetCode, *output.Message)
		return fmt.Errorf(*output.Message)
	}
	klog.Infof("Call IaaS DeleteVolume %s succeed", id)
	return nil
}

// AttachVolume
// 1. attach volume on instance
// 2. wait job
func (q *qingCloudManager) AttachVolume(volId string, instId string) error {
	// set input parameter
	input := &qcservice.AttachVolumesInput{
		Volumes:  []*string{&volId},
		Instance: &instId,
	}
	// attach volume
	klog.Infof("Call IaaS AttachVolume request volume id: %s, instance id: %s, zone: %s", volId, instId,
		q.GetZone())
	output, err := q.volumeService.AttachVolumes(input)
	if err != nil {
		return err
	}
	// check output
	if *output.RetCode != 0 {
		klog.Errorf("Ret code: %d, message: %s", *output.RetCode, *output.Message)
		return fmt.Errorf(*output.Message)
	}
	// wait job
	klog.Infof("Call IaaS WaitJob %s", *output.JobID)
	if err := q.waitJob(*output.JobID); err != nil {
		return err
	}
	// check output
	if *output.RetCode != 0 {
		klog.Errorf("Ret code: %d, message: %s", *output.RetCode, *output.Message)
		return fmt.Errorf(*output.Message)
	}
	klog.Infof("Call IaaS AttachVolume %s on instance %s succeed", volId, instId)
	return nil
}


func (q *qingCloudManager) NodeAttachVolume(volId string) error {
	return nil
}

func (q *qingCloudManager) NodeDetachVolume(volId string) error{
	return nil
}

func (q *qingCloudManager) NodeGetDevice(volId string) (string,error) {
	return "", nil
}
// detach volume
// 1. detach volume
// 2. wait job
func (q *qingCloudManager) DetachVolume(volumeId string, instanceId string) error {
	// set input parameter
	input := &qcservice.DetachVolumesInput{
		Volumes:  []*string{&volumeId},
		Instance: &instanceId,
	}
	// detach volume
	klog.Infof("Call IaaS DetachVolume request volume id: %s, instance id: %s, zone: %s", volumeId,
		instanceId, q.GetZone())
	output, err := q.volumeService.DetachVolumes(input)
	if err != nil {
		return err
	}
	// check output
	if *output.RetCode != 0 {
		klog.Errorf("Ret code: %d, message: %s", *output.RetCode, *output.Message)
		return fmt.Errorf(*output.Message)
	}
	// wait job
	klog.Infof("Call IaaS WaitJob %s", *output.JobID)
	if err := q.waitJob(*output.JobID); err != nil {
		return err
	}
	// check output
	if *output.RetCode != 0 {
		klog.Errorf("Ret code: %d, message: %s", *output.RetCode, *output.Message)
		return fmt.Errorf(*output.Message)
	}
	klog.Infof("Call IaaS DetachVolume %s succeed", volumeId)
	return nil
}

// ResizeVolume can expand the size of a volume offline
// requestSize: GB
func (q *qingCloudManager) ResizeVolume(volumeId string, requestSize int) error {
	// resize
	klog.Infof("Call IaaS ResizeVolume request volume %s size %d Gib in zone [%s]",
		volumeId, requestSize, q.GetZone())
	input := &qcservice.ResizeVolumesInput{
		Size:    &requestSize,
		Volumes: []*string{&volumeId},
	}
	output, err := q.volumeService.ResizeVolumes(input)
	if err != nil {
		return err
	}
	// check output
	if *output.RetCode != 0 {
		klog.Errorf("Ret code: %d, message: %s", *output.RetCode, *output.Message)
		return errors.New(*output.Message)
	}
	// wait job
	klog.Infof("Call IaaS WaitJob %s", *output.JobID)
	if err := q.waitJob(*output.JobID); err != nil {
		return err
	}
	// check output
	if *output.RetCode != 0 {
		klog.Errorf("Ret code: %d, message: %s", *output.RetCode, *output.Message)
		return errors.New(*output.Message)
	}
	klog.Infof("Call IaaS ResizeVolume id %s size %d succeed", volumeId, requestSize)
	return nil
}

// CloneVolume clones a volume
// Return:
//   volume id, nil: succeed to clone volume and return volume id
//   nil, error: failed to clone volume
func (q *qingCloudManager) CloneVolume(volName string, volType int, srcVolId string, zone string) (newVolId string,
	err error) {
	// 0. Set CreateVolume args
	// create volume count
	count := 1
	input := &qcservice.CloneVolumesInput{
		Count:      &count,
		Volume:     &srcVolId,
		VolumeName: &volName,
		VolumeType: &volType,
		Zone:       &zone,
	}
	// 1. Clone volume
	klog.Infof("Call IaaS CloneVolume request name: %s, source volume id: %s, zone: %s", volName, srcVolId, zone)
	output, err := q.volumeService.CloneVolumes(input)
	if err != nil {
		return "", err
	}
	// wait job
	klog.Infof("Call IaaS WaitJob %s", *output.JobID)
	if err := q.waitJob(*output.JobID); err != nil {
		return "", err
	}
	// check output
	if *output.RetCode != 0 {
		klog.Errorf("Ret code: %d, message: %s", *output.RetCode, *output.Message)
		return "", errors.New(*output.Message)
	}
	newVolId = *output.Volumes[0]
	klog.Infof("Call IaaS CloneVolume name %s id %s succeed", volName, newVolId)
	return newVolId, nil
}

// Find instance by instance ID
// Return: 	nil,	nil: 	not found instance
//			instance, nil: 	found instance
//			nil, 	error:	internal error
func (q *qingCloudManager) FindInstance(id string) (instance *qcservice.Instance, err error) {
	seeAppCluster := cloud.EnableDescribeInstanceAppCluster
	verboseMode := cloud.EnableDescribeInstanceVerboseMode
	// set describe instance input
	input := qcservice.DescribeInstancesInput{
		Instances:     []*string{&id},
		IsClusterNode: &seeAppCluster,
		Verbose:       &verboseMode,
	}
	// call describe instance
	output, err := q.instanceService.DescribeInstances(&input)
	// error
	if err != nil {
		return nil, err
	}
	if *output.RetCode != 0 {
		klog.Errorf("Ret code: %d, message: %s", *output.RetCode, *output.Message)
		return nil, fmt.Errorf(*output.Message)
	}
	// not found instances
	switch *output.TotalCount {
	case 0:
		return nil, nil
	case 1:
		if *output.InstanceSet[0].Status == cloud.InstanceStatusCreased || *output.InstanceSet[0].Status == cloud.InstanceStatusTerminated {
			return nil, nil
		}
		return output.InstanceSet[0], nil
	default:
		return nil, fmt.Errorf("find duplicate instances id %s in %s", id, q.instanceService.Config.Zone)
	}
}

// GetZone
// Get current zone in Qingcloud IaaS
func (q *qingCloudManager) GetZone() string {
	if q == nil {
		return ""
	}
	return q.cloudService.Config.Zone
}

// GetZoneList gets active zone list
func (q *qingCloudManager) GetZoneList() (zones []string, err error) {
	output, err := q.cloudService.DescribeZones(&qcservice.DescribeZonesInput{})
	// Error:
	// 1. Error is not equal to nil.
	if err != nil {
		return nil, err
	}
	if output == nil {
		klog.Error("should not response nil")
	}
	for i := range output.ZoneSet {
		if *output.ZoneSet[i].Status == cloud.ZoneStatusActive {
			zones = append(zones, *output.ZoneSet[i].ZoneID)
		}
	}
	return zones, nil
}

// FindTags finds and gets tags information
func (q *qingCloudManager) FindTag(tagId string) (tagInfo *qcservice.Tag, err error) {
	if len(tagId) == 0 {
		return nil, nil
	}
	input := &qcservice.DescribeTagsInput{
		Tags: []*string{&tagId},
	}

	output, err := q.tagService.DescribeTags(input)
	if err != nil {
		return nil, err
	}
	if *output.RetCode != 0 {
		klog.Errorf("Ret code: %d, message: %s", *output.RetCode, *output.Message)
		return nil, fmt.Errorf("call IaaS DescribeTags err: tag id %s in %s", tagId, q.GetZone())
	}
	switch *output.TotalCount {
	// Not found tag
	case 0:
		return nil, nil
	// Found one tag
	case 1:
		return output.TagSet[0], nil
	// Found duplicate tags
	default:
		return nil, fmt.Errorf("call IaaS DescribeTags err: find duplicate tags, tag id %s in %s",
			tagId, q.GetZone())
	}
}

// AttachTag adds a slice of tags on a object
func (q *qingCloudManager) AttachTags(tagsId []string, resourceId string, resourceType string) (err error) {
	if len(tagsId) == 0 {
		klog.Infof("No tags need attached")
		return nil
	}
	var tagPairs []*qcservice.ResourceTagPair
	for index := range tagsId {
		tagPairs = append(tagPairs, &qcservice.ResourceTagPair{
			ResourceID:   &resourceId,
			ResourceType: &resourceType,
			TagID:        &tagsId[index],
		})
	}
	input := &qcservice.AttachTagsInput{tagPairs}
	output, err := q.tagService.AttachTags(input)
	if err != nil {
		return err
	}
	if *output.RetCode != 0 {
		klog.Errorf("Ret code: %d, message: %s", *output.RetCode, *output.Message)
		return fmt.Errorf("call IaaS AttachTags err: tag id %v, resource id %s, resource type %s in %s", tagsId,
			resourceId, resourceType, q.GetZone())
	}
	klog.Infof("Call IaaS AttachTags %v on resource %s succeed", tagsId, resourceId)
	return nil
}

func (q *qingCloudManager) Probe() error {
	zones , err := q.GetZoneList()
	if err != nil{
		return err
	}
	klog.V(5).Infof("get active zone lists [%v]", zones)
	return nil
}

func (q *qingCloudManager) GetTopology(instanceId string) (*csi.Topology, error) {
	instInfo, err := q.FindInstance(instanceId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if instInfo == nil {
		return nil, status.Errorf(codes.NotFound, "cannot found instance %s", instanceId)
	}

	//instanceType, ok := driver.InstanceTypeName[driver.InstanceType(*instInfo.InstanceClass)]
	//if !ok {
	//	return nil, status.Errorf(codes.InvalidArgument, "unsupported instance type %d", *instInfo.InstanceClass)
	//}
	top := &csi.Topology{
		Segments: map[string]string{
			//ns.driver.TopologyInstanceTypeKey(): instanceType,
			//ns.driver.GetTopologyZoneKey():   *instInfo.ZoneID,
		},
	}
	return top,nil
}
// IsValidTags checks tags exists.
func (q *qingCloudManager) IsValidTags(tagsId []string) bool {
	for _, tagId := range tagsId {
		tagInfo, err := q.FindTag(tagId)
		if err != nil {
			return false
		}
		if tagInfo == nil {
			return false
		}
	}
	return true
}

func (q *qingCloudManager) waitJob(jobId string) error {
	err := qcclient.WaitJob(q.jobService, jobId, cloud.WaitJobTimeout, cloud.WaitJobInterval)
	if err != nil {
		return fmt.Errorf("call IaaS WaitJob id %s, error: ", err)
	}
	return nil
}
