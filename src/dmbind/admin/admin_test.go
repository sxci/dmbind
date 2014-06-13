package admin

import (
	"testing"
	"dmbind/lib/context"
)

func init() {
	Setup("ikbear", "qiniutek", 10)
}

func Test(t *testing.T) {
	return
	path := "/?"
	ctx := context.NewFakeContext("GET", path, nil)
	List(ctx)
	t.Error(ctx.FakeString())
}
