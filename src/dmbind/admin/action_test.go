package admin

import (
	"net/url"
	"testing"
	"dmbind/lib/context"
)

func TestStatus(t *testing.T) {
	return
	err := context.Call(nil, Status, url.Values {
		"id": {"52009a96f25228548c000002"},
		"status": {"0"},
	})
	if err != nil { t.Fatal(err) }
}
