package test

import (
	"net/url"
	"dmbind/lib/context"
	"dmbind/lib/pub"
	"net/http"
)

func Curl(c *context.Context) {
	host := c.MustString("host")
	testUrl := c.MustString("testUrl")
	req, _ := http.NewRequest("GET", testUrl, nil)
	req.Host = host
	client := &http.Client{}
	resp, err := client.Do(req)
	c.ReplyIfError(err)
	c.ReplyObj(resp.StatusCode)
}

func Check(c *context.Context) {
	host := c.MustString("host")
	testUrl := c.MustString("testUrl")
	bucket := c.MustString("bucket")
	
	err := pub.Publish(host, bucket)
	c.ReplyIfError(err)
	defer pub.Unpublish(host)
	var ret int
	err = context.Call(&ret, Curl, url.Values {
		"host": {host},
		"testUrl": {testUrl},
	})
	c.ReplyIfError(err)
	c.ReplyObj(ret)
}
