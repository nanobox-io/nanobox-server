package util_test

import "testing"
import "github.com/nanobox-io/nanobox-server/util"

func TestLock(t *testing.T) {
	util.Lock()
	if util.LockCount() != 1 {
		t.Errorf("the lock count should be 1 but it is %d", util.LockCount())
	}
	util.Unlock()
	if util.LockCount() != 0 {
		t.Errorf("the lock count should be 0 but it is %d", util.LockCount())
	}
}
