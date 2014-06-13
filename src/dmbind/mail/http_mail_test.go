package mail

import (
	"net/url"
	"testing"
	"dmbind/lib/context"
)

func TestSend(t *testing.T) {
	err := context.Call(nil, Send, url.Values {
		"to": {"me@chenye.org"},
		"subject": {"hello subject"},
		"content": {`<html><form action="http://localhost:8080/"><input value="df" /><input type="submit"/></form></html>`},
	})
	if err != nil {
		t.Fatal(err)
	}
}
