package api

type ResponseHeader struct {
	Op      string `json:"op"`
	RetCode int    `json:"ret_code"`
	Reason  string `json:"reason"`
}

func (r *ResponseHeader) Header() *ResponseHeader {
	return r
}

type Response interface {
	Header() *ResponseHeader
}

type CreateVolumeRequest struct {
	Op       string `json:"op"`
	PoolName string `json:"pool_name"`
	Name     string `json:"name"`
	Size     int64  `json:"size"`
	RepCount int    `json:"rep_count"`
}

type CreateVolumeResponse struct {
	ResponseHeader
	Id   int `json:"id"`
	Size int `json:"size"`
}

type DeleteVolumeRequest struct {
	Op       string `json:"op"`
	PoolName string `json:"pool_name"`
	Name     string `json:"name"`
}

type DeleteVolumeResponse struct {
	ResponseHeader
	Id int `json:"id"`
}

type ListVolumeRequest struct {
	Op       string `json:"op"`
	PoolName string `json:"pool_name"`
	Name     string `json:"name"`
}

type ListVolumeResponse struct {
	ResponseHeader
	Count   int      `json:"count"`
	Volumes []Volume `json:"volumes"`
}
