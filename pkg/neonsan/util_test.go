package neonsan

import "testing"

func TestFormatVolumeSize(t *testing.T) {
	tests := []struct {
		name      string
		inputSize int64
		step      int64
		outSize   int64
	}{
		{
			name:      "Format 4Gi, step 1Gi",
			inputSize: 4294967296,
			step:      gib,
			outSize:   4294967296,
		},
		{
			name:      "Format 4Gi, step 10Gi",
			inputSize: 4294967296,
			step:      gib * 10,
			outSize:   gib * 10,
		},
		{
			name:      "Format 4Gi, step 3Gi",
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
