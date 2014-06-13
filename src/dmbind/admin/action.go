package admin

import (
	"errors"
	"strconv"
	"net/url"
	"encoding/base64"
	"dmbind/lib/context"
)

func Status(c *context.Context) {
	status := c.MustString("status")
	ids := c.Strings("id", nil)
	domains := c.Strings("domain", nil)
	from := c.String("from", "")
	cdn := c.String("cdn", "")

	if len(domains) > 0 {
		var list []string
		if err := context.Call(&list, List, url.Values {
			"domain": domains,
			"field": {"id"},
			"showList": {"1"},
			"status": {from},
			"cdn": {cdn},
		}); err != nil {
			c.ReplyError(err)
			return
		}

		newIds := make([]string, len(ids)+len(list))
		copy(newIds[:len(ids)], ids)
		length := len(ids)
		for _, listId := range list {
			isExist := false
			for _, id := range newIds[:len(ids)]{
				if id == listId {
					isExist = true
					break
				}
			}
			if isExist { continue }
			newIds[length] = listId
			length ++
		}
		ids = newIds[:length]
	}

	defer func() {
		cacheList = nil
	}()

	if len(ids) <= 0 && len(domains) > 0 {
		c.ReplyErrorInfo("error! has domain " + domains[0] + " and got empty ids")
		return
	}
	for _, id := range ids {
		_, err := postAdmin("/proxy/api?biz/admin/cdomain/set_status", url.Values {
			"id": {id},
			"status": {status},
		})
		if err != nil {
			c.ReplyErrorInfo("set id " + id + " to status " + status + " fail: " + err.Error())
			return
		}
		c.Info("set id " + id + " to status " + status, "success!", "domain", domains)
	}
	c.ReplyObj(true)
}

func PassId(c *context.Context) {
	id := c.MustString("id")
	domain := c.MustString("domain")
	bucket := c.MustString("bucket")
	uid := c.MustString("uid")
	c.ReplyIfError(PubBucketDomain(domain, bucket, uid))
	c.Info("pub domain:", domain, "bucket:", bucket, "uid:", uid)
	c.ReplyIfError(context.Call(nil, Status, url.Values {
		"id": {id},
		"status": {"2"},
	}))
}

// 1. publish domain
// 2. update status
// 3. send mail
func Pass(c *context.Context) {
	domains := c.MustStrings("domain")
	from := c.String("from", "3")
	mail := c.Bool("mail", true)
	cdn := c.String("cdn", "")
	
	var rets []DomainInfo
	c.ReplyIfError(context.Call(&rets, List, url.Values {
		"domain": domains,
		"status": {from},
		"cdn": {cdn},
	}))

	for _, info := range rets {
		c.ReplyIfError(PubBucketDomain(info.Domain, info.Bucket, strconv.Itoa(info.Uid)))
		c.Info("pub domain:", info.Domain, "bucket:", info.Bucket, "uid:", info.Uid)
	}
	c.ReplyIfError(context.Call(nil, Status, url.Values {
		"domain": domains,
		"from": {from},
		"status": {"2"},
		"cdn": {cdn},
	}))
	if mail {
		c.ReplyIfError(context.Call(nil, Report, url.Values {
			"domain": domains,
			"mail": {"1"},
		}))
	}
	c.ReplyObj(true)
	return
}

func PubBucketDomain(domain, bucket, uid string) (err error) {
	u := "/proxy/api?rs/admin/publish/" + encode(domain)
	u += "/from/" + bucket
	u += "/uid/" + uid
	_, err = postAdmin(u, nil)
	if err != nil {
		info := "publish failure:" + domain + "," + bucket + "," + uid
		info += ",detail:" + err.Error()
		err = errors.New(info)
	}
	return
}

func encode(u string) string {
	return base64.URLEncoding.EncodeToString([]byte(u))
}

// 1. update status
// 2. send mail
func Deny(c *context.Context) {
	domains := c.MustStrings("domain") // host
	from := c.String("from", "")
	mail := c.Bool("mail", true)
	c.ReplyIfError(context.Call(nil, Status, url.Values {
		"domain": domains,
		"status": {"0"},
		"from": {from},
	}))
	if mail {
		c.ReplyIfError(context.Call(nil, Report, url.Values {
			"domain": domains,
			"mail": {"1"},
		}))
	}
	c.ReplyObj(true)
	return
}
