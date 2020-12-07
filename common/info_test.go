package common

import "testing"

func TestNewBasicInfo(t *testing.T) {
	binfo := NewBasicInfo()
	t.Logf("%#v", binfo)
}
