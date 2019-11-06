/*
Copyright 2018 Yunify, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package common

import (
	"fmt"
	"strconv"

	//"github.com/container-storage-interface/spec/lib/go/csi/v0"
	"os/exec"
	"time"
)

// ExecCommand
// Return cases:	normal output,	nil:	normal output
//					error logs,		error:	command execute error
func ExecCommand(command string, args []string) ([]byte, error) {
	time.Sleep(time.Second)
	cmd := exec.Command(command, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("execute cmd [%s] args [%v] error: code [%s], msg [%s]",
			command, args, err.Error(), output)
	}
	return output, nil
}

// ParseIntToDec convert number string to decimal number string
func ParseIntToDec(hex string) (dec string) {
	i64, err := strconv.ParseInt(hex, 0, 64)
	if err != nil {
		return ""
	}
	return strconv.FormatInt(i64, 10)
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
