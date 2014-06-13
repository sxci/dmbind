package dnspod

import "testing"
import "dmbind/lib/context"
import "net/url"

func TestCName(t *testing.T) {
	err := context.Call(nil, DnspodCName, url.Values {
		"bucket": {"weibo"},
	})
	if err != nil { t.Fatal(err) }
}
