package neonsan

type neonsanStorageClass struct {
	Replicas	int `json:"replicas"`
	VolumeFsType	string `json:"fsType"`
	Pool	string `json:"pool"`
}

