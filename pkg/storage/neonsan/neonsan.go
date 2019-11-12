package neonsan

import (
	"github.com/yunify/qingstor-csi/pkg/storage"
)

type neonsan struct {
	confFile string
	poolName string
}

func New(confFile, poolName string) storage.Provider {
	return &neonsan{
		confFile: confFile,
		poolName: poolName,
	}
}
