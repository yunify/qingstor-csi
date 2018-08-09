package neonsan

import (
	"encoding/json"
	"fmt"
	"github.com/container-storage-interface/spec/lib/go/csi/v0"
	"github.com/golang/glog"
	"os"
	"os/exec"
	"path"
)

const (
	Int64Max     = int64(^uint64(0) >> 1)
	PluginFolder = "/var/lib/kubelet/plugins/"
)

const (
	kib    int64 = 1024
	mib    int64 = kib * 1024
	gib    int64 = mib * 1024
	gib100 int64 = gib * 100
	tib    int64 = gib * 1024
	tib100 int64 = tib * 100
)

const (
	FileSystemExt3    string = "ext3"
	FileSystemExt4    string = "ext4"
	FileSystemXfs     string = "xfs"
	FileSystemDefault string = FileSystemExt4
)

var (
	ConfigFilePath string = "/etc/neonsan/qbd.conf"
)

// ExecCommand
// Return cases:	normal output,	nil:	normal output
//					error logs,		error:	command execute error
func ExecCommand(command string, args []string) ([]byte, error) {
	glog.Infof("execCommand: command = \"%s\", args = \"%v\"", command, args)
	cmd := exec.Command(command, args...)
	output, err := cmd.CombinedOutput()
	if err != nil{
		return nil, fmt.Errorf("code [%s]: message [%s]", err.Error(), output)
	}
	return output, nil
}

// ContainsVolumeCapability
// Does Array of VolumeCapability_AccessMode contain the volume capability of subCaps
func ContainsVolumeCapability(accessModes []*csi.VolumeCapability_AccessMode, subCaps *csi.VolumeCapability) bool {
	for _, cap := range accessModes {
		if cap.GetMode() == subCaps.GetAccessMode().GetMode() {
			return true
		}
	}
	return false
}

// ContainsVolumeCapabilities
// Does array of VolumeCapability_AccessMode contain volume capabilities of subCaps
func ContainsVolumeCapabilities(accessModes []*csi.VolumeCapability_AccessMode, subCaps []*csi.VolumeCapability) bool {
	for _, v := range subCaps {
		if !ContainsVolumeCapability(accessModes, v) {
			return false
		}
	}
	return true
}

// FormatVolumeSize convert volume size properly
func FormatVolumeSize(inputSize int64, step int64) int64 {
	if inputSize <= gib || step < 0 {
		return gib
	}
	if inputSize%step != 0 {
		return inputSize + gib
	}
	return inputSize
}

// Check file system type
// Support: ext3, ext4 and xfs
func IsValidFileSystemType(fs string) bool {
	switch fs {
	case FileSystemExt3:
		return true
	case FileSystemExt4:
		return true
	case FileSystemXfs:
		return true
	default:
		return false
	}
}

//	CreatePersistentStorage create path to save volume info files
func CreatePersistentStorage(persistentStoragePath string) error {
	if _, err := os.Stat(persistentStoragePath); os.IsNotExist(err) {
		if err := os.MkdirAll(persistentStoragePath, os.FileMode(0755)); err != nil {
			return err
		}
	} else {
	}
	return nil
}

//	PersistVolInfo save volume info
func PersistVolInfo(image string, persistentStoragePath string, volInfo *volumeInfo) error {
	file := path.Join(persistentStoragePath, image+".json")
	fp, err := os.Create(file)
	if err != nil {
		glog.Errorf("failed to create persistent storage file %s with error: %v\n", file, err)
		return fmt.Errorf("rbd: create err %s/%s", file, err)
	}
	defer fp.Close()
	encoder := json.NewEncoder(fp)
	if err = encoder.Encode(volInfo); err != nil {
		glog.Errorf("failed to encode volInfo: %+v for file: %s with error: %v\n", volInfo, file, err)
		return fmt.Errorf("encode err: %v", err)
	}
	glog.Infof("successfully saved volInfo: %+v into file: %s\n", volInfo, file)
	return nil
}

//	LoadVolInfo load volume info
func LoadVolInfo(image string, persistentStoragePath string, volInfo *volumeInfo) error {
	file := path.Join(persistentStoragePath, image+".json")
	fp, err := os.Open(file)
	if err != nil {
		return fmt.Errorf("open err %s/%s", file, err)
	}
	defer fp.Close()

	decoder := json.NewDecoder(fp)
	if err = decoder.Decode(volInfo); err != nil {
		return fmt.Errorf("decode err: %v", err)
	}

	return nil
}

//	DeleteVolInfo delete volume info
func DeleteVolInfo(image string, persistentStoragePath string) error {
	file := path.Join(persistentStoragePath, image+".json")
	glog.Infof("deleting file for Volume: %s at: %s resulting path: %+v\n", image, persistentStoragePath, file)
	err := os.Remove(file)
	if err != nil {
		if err != os.ErrNotExist {
			return fmt.Errorf("error removing file: %s/%s", file, err)
		}
	}
	return nil
}
