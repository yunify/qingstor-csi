package neonsan

import "fmt"

// poolInfo: stats pool
// total, free, used: pool size in bytes
type poolInfo struct {
	id    string
	name  string
	total int64
	free  int64
	used  int64
}

//	FindPool
// 	Description:	get pool detail information
//	Input:	pool name:	string
//	Return cases:	pool,	nil:	found pool
//					nil,	nil:	pool not found
//					nil,	err:	error
func FindPool(poolName string) (outPool *poolInfo, err error) {
	args := []string{"stats_pool", "-pool", poolName, "-c", ConfigFilePath}
	output, err := ExecCommand(CmdNeonsan, args)
	if err != nil {
		return nil, err
	}
	outPool = ParsePoolInfo(string(output))
	if outPool == nil {
		return nil, nil
	}
	if outPool.name != poolName {
		return nil, fmt.Errorf("mismatch pool name: expect %s, but actually %s", poolName, outPool.name)
	}
	return outPool, nil
}
