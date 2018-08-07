package neonsan

import (
	"os/exec"
	"github.com/golang/glog"
)

const (
	Int64Max             = int64(^uint64(0) >> 1)
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

var(
	ConfigFilePath string = "/etc/neonsan/qbd.conf"
)

func execCommand(command string, args []string) ([]byte, error) {
	glog.Infof("execCommand: command = \"%s\", args = \"%v\"", command, args)
	cmd := exec.Command(command, args...)
	return cmd.CombinedOutput()
}