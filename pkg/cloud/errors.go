package cloud

import (
	"regexp"
)

var (
	leaseInfoNotReadyRegex    = regexp.MustCompile(`QingCloud Error: Code \(1400\), Message \(PermissionDenied, resource \[[a-z]{1,3}-\w{8}] lease info not ready yet, please try later\)`)
	snapshotNotAvailableRegex = regexp.MustCompile(`QingCloud Error: Code \(1400\), Message \(PermissionDenied, snapshot \[[a-z]{1,3}-\w{8}] is not available, can not create volume from it\)`)
	tryLaterRegex             = regexp.MustCompile(`please try later`)
)

func IsLeaseInfoNotReady(err error) bool {
	return leaseInfoNotReadyRegex.MatchString(err.Error())
}

func IsSnapshotNotAvailable(err error) bool {
	return snapshotNotAvailableRegex.MatchString(err.Error())
}

func IsTryLater(err error) bool {
	return tryLaterRegex.MatchString(err.Error())
}
