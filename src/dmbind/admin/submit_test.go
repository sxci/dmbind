package admin

import (
	"testing"
	"dmbind/lib/context"
)

func TestFilterSubmit(t *testing.T) {
	return
	var a interface{}
	err := context.Call(&a, FilterSubmit, nil)
	if err != nil { t.Fatal(err) }
	t.Error(a)
}
