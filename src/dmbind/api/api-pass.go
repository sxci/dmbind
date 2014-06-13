package api

import (
	"net/url"
	"dmbind/lib/context"
)

func CheckAndPass(c *context.Context) {
	domain := c.MustString("domain")
	var result bool
	err := context.Call(&result, TestCheckDomain, url.Values {
		"domain": {domain},
	})
	c.ReplyIfError(err)
	var ret interface{}
	err = context.Call(&ret, AdminPassDomain, url.Values {
		"domain": {domain},
	})
	c.ReplyIfError(err)
	c.ReplyObj(ret)
}
