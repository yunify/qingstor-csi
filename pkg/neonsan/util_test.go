package neonsan

import (
	"testing"
)

func TestFormatVolumeSize(t *testing.T) {
	tests := []struct {
		name      string
		inputSize int64
		step      int64
		outSize   int64
	}{
		{
			name:      "format 4Gi, step 1Gi",
			inputSize: 4294967296,
			step:      gib,
			outSize:   4294967296,
		},
		{
			name:      "format 4Gi, step 10Gi",
			inputSize: 4294967296,
			step:      gib * 10,
			outSize:   gib * 10,
		},
		{
			name:      "format 4Gi, step 3Gi",
			inputSize: 4294967296,
			step:      gib * 3,
			outSize:   gib * 6,
		},
	}
	for _, v := range tests {
		out := FormatVolumeSize(v.inputSize, v.step)
		if v.outSize != out {
			t.Errorf("name %s: expect %d, but actually %d", v.name, v.outSize, out)
		}
	}
}

func TestParseIntToDec(t *testing.T) {
	tests := []struct {
		name string
		hex  string
		dec  string
	}{
		{
			name: "success parse",
			hex:  "0x3ff7000000",
			dec:  "274726912000",
		},
		{
			name: "failed parse",
			hex:  "321",
			dec:  "321",
		},
	}
	for _, v := range tests {
		ret := ParseIntToDec(v.hex)
		if v.dec != ret {
			t.Errorf("name [%s]: expect [%s], but actually [%s]", v.name, v.dec, ret)
		}

	}
}
