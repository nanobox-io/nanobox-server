package util

import (
	"sync/atomic"
)

var lockCount = int64(0)

func LockCount() int {
	return int(lockCount)
}

func Lock() {
	atomic.AddInt64(&lockCount, 1)
}

func Unlock() {
	atomic.AddInt64(&lockCount, -1)
}
