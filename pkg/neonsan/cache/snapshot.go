package cache

import (
	"fmt"
	"github.com/golang/glog"
	"github.com/yunify/qingstor-csi/pkg/neonsan/manager"
	"reflect"
	"sort"
)

type SnapshotCache interface {
	// Add snapshot into map
	// 1. snapshot name does not exist, add snapshot information normally.
	// 2. snapshot name exists but snapshot info is not equal to input
	// snapshot info, add snapshot failed.
	// 3. snapshot name exists and snapshot info is equal to input snapshot
	// info, add snapshot succeed.
	Add(info *manager.SnapshotInfo) bool
	// Find snapshot information by snapshot name
	// If founded snapshot, return snapshot info
	// If not founded snapshot, return nil
	Find(snapName string) *manager.SnapshotInfo
	// Delete snapshot information form map
	Delete(snapName string)
	// Add all snapshot information into map
	Sync() error
	// List all snapshot info
	List() []*manager.SnapshotInfo
}

type SnapshotCacheType struct {
	Snaps map[string]*manager.SnapshotInfo
}

var _ SnapshotCache = &SnapshotCacheType{}

func (snapCache *SnapshotCacheType) New() {
	// key-value: name-info
	snapCache.Snaps = make(map[string]*manager.SnapshotInfo)
}

func (snapCache *SnapshotCacheType) Add(info *manager.SnapshotInfo) bool {
	if info == nil {
		return false
	}
	if exInfo, ok := snapCache.Snaps[info.Name]; ok {
		// already exist
		if reflect.DeepEqual(info, exInfo) {
			// new info == exist info
			return true
		} else {
			// new info != exist info
			return false
		}
	}
	// not exist
	snapCache.Snaps[info.Name] = info
	return true
}

func (snapCache *SnapshotCacheType) Find(snapName string) *manager.SnapshotInfo {
	if exInfo, ok := snapCache.Snaps[snapName]; ok {
		// already exist
		return exInfo
	}
	// not exist
	return nil
}

func (snapCache *SnapshotCacheType) Delete(snapName string) {
	if _, ok := snapCache.Snaps[snapName]; ok {
		// already exist
		delete(snapCache.Snaps, snapName)
	}
}

func (snapCache *SnapshotCacheType) Sync() (err error) {
	// get full snapshot list
	for _, v := range manager.ListPoolName() {
		// visit each pool
		vols, err := manager.ListVolumeByPool(v)
		if err != nil {
			return err
		}
		for _, volInfo := range vols {
			// visit each volume
			glog.Info(volInfo)
			volSnapList, err := manager.ListSnapshotByVolume(volInfo.Name, volInfo.Pool)
			glog.Info(volSnapList)
			if err != nil {
				return err
			}
			for i := range volSnapList {
				if snapCache.Add(volSnapList[i]) {
					glog.Infof("add snapshot [%s] into cache successfully", volSnapList[i].Name)
				} else {
					return fmt.Errorf("add snapshot [%v] failed, already exits but incompatiably", volSnapList[i])
				}
			}
		}
	}
	return nil
}

// List: list snapshot map by snapshot name
func (snapCache *SnapshotCacheType) List() (list []*manager.SnapshotInfo) {
	sortedKeys := make([]string, 0)
	for k := range snapCache.Snaps {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Strings(sortedKeys)
	for _, name := range sortedKeys {
		list = append(list, snapCache.Snaps[name])
	}
	return list
}
