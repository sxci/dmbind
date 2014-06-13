package api

import (
	"time"
	"strconv"
	"net/url"
	"errors"
	"dmbind/lib/context"
	"dmbind/lib/pub"
	"dmbind/test"
	"dmbind/admin"
	"math/rand"
)

var testBucket = "demo"
var testUrl = "http://dmbind.test.qiniudn.com/"

func init() {
	rand.Seed(time.Now().Unix())
}

type CdnResult struct { Wangsu,Tongxing,Lanxun bool }

func getDomainCdnList(domain, status string) (ret CdnResult, err error) {
	var cdnList []int
	err = context.Call(&cdnList, admin.List, url.Values {
		"status": {status},
		"domain": {domain},
		"field": {"cdn"},
		"showList": {"1"},
	})
	if err == nil && len(cdnList) == 0 {
		err = errors.New("record which domain specified not found")
	}
	if err != nil { return }
	for _, i := range cdnList {
		switch i {
		case 0:
			ret.Wangsu = true
		case 1:
			ret.Tongxing = true
		}
	}
	return
}

func makeTestHost(host string) string {
	return "dmbind.test." + host
}

func TestBenchDomain(c *context.Context) {
	domain := c.MustString("domain")
	cdnRet, err := getDomainCdnList(domain, "3")
	c.ReplyIfError(err)
	result := true
	if cdnRet.Wangsu || cdnRet.Tongxing {
		var success bool
		c.ReplyIfError(context.Call(&success, test.Alibench, url.Values {
			"domain": {domain},
			"bucket": {testBucket},
			"testUrl": {"dmbind."+strconv.Itoa(rand.Int() % 1000)+".qiniudn.com"},
		}))
		result = result && success
	}
	c.ReplyObj(result)
}

func TestCheckDomain(c *context.Context) {
	domain := c.MustString("domain")
	cdnRet, err := getDomainCdnList(domain, "3")
	c.ReplyIfError(err)
	result := true
	testDomain := makeTestHost(domain)
	pub.Publish(testDomain, testBucket)
	if cdnRet.Wangsu || cdnRet.Tongxing {
		var code int
		c.ReplyIfError(context.Call(&code, test.Check, url.Values {
			"host": {testDomain},
			"testUrl": {testUrl},
			"bucket": {testBucket},
		}))
		if code == 503 { result = false }
	} else {
		result = false
	}
	c.ReplyObj(result)
}
