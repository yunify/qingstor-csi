package neonsan

import "testing"

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
			errStr: "-bash: fake: command not found",
		},
		{
			name:   "error pwd -as",
			cmd:    "pwd",
			args:   []string{"-as"},
			errStr: "-bash: fake: command not found",
		},
	}
	for _, v := range tests {
		bytes, err := ExecCommand(v.cmd, v.args)
		t.Logf("name %s: output [%s], error string [%v]", v.name, bytes, err)
	}
}
