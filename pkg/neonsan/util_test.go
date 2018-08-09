package neonsan

import (
	"testing"
	"strings"
)

func TestExecCommand(t *testing.T) {
	tests := []struct {
		name   string
		cmd    string
		args   []string
		errStr string
	}{
		{
			name:   "normal ls -al",
			cmd:    "ls",
			args:   []string{"-al"},
			errStr: "",
		},
		{
			name:   "error fake",
			cmd:    "fake",
			args:   nil,
			errStr: "command not found",
		},
		{
			name:   "error pwd -as",
			cmd:    "pwd",
			args:   []string{"-as"},
			errStr: "pwd: invalid option -- 'a'",
		},
	}
	for _, v := range tests {
		_, err := ExecCommand(v.cmd, v.args)
		if v.errStr == "" && err == nil{
			continue
		}else if v.errStr != "" && err != nil{
			if !strings.Contains(err.Error(), v.errStr){
				t.Errorf("name %s: expect error [%s], but actually [%s]", v.name, v.errStr, err.Error())
			}
		}else{
			t.Errorf("name %s: expect error [%s], but actually [%v]", v.name, v.errStr, err)
		}
	}
}
