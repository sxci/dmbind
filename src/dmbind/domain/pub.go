package domain

import (
	"dmbind/lib/context"
	"dmbind/lib/pub"
)

func Pub(c *context.Context) {
	domain := c.MustString("domain")
	bucket := c.String("bucket", "demo")
	err := pub.Publish(domain, bucket)
	c.ReplyIfError(err)
	c.ReplyObj(true)
}

func Unpub(c *context.Context) {
	domain := c.MustString("domain")
	err := pub.Unpublish(domain)
	c.ReplyIfError(err)
	c.ReplyObj(true)
}
