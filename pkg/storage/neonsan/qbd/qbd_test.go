package qbd

import "testing"

func TestInfoVolume(t *testing.T) {
	input := `
volume:                  kube/happy
config:                  /etc/neonsan/qbd.conf
volume id:               0xb4000000
transport type:          tcp
status:                  opened, normal
device:                  /dev/qbd0
shards information([+]:connected [-]:closed [@]:closing [x]:failed [?]:unknown):
shard[0]:
IP[0]:192.168.0.4:7800 [+]
`
	device, err := parseDevice(input)
	if err != nil {
		t.Error(err)
	}
	t.Log(device)
}
