package stat

import (
	"net/url"
	"dmbind/lib/context"
	"dmbind/domain"
	"dmbind/admin"
)

type Txws struct {
	Qiniudn string `json:"qiniudn"`
	Domain string `json:"domain,omitempty"`
}

func TxWs(c *context.Context) {
	var buckets []string
	err := context.Call(&buckets, domain.TxWs, nil)
	c.ReplyIfError(err)
	var ret []admin.DomainInfo
	err = context.Call(&ret, admin.List, url.Values {
		"bucket": buckets,
		"status": {"2", "4"},
		"field": {"domain-bucket"},
	})
	r := make([]Txws, len(buckets))
	bucketDomain := make(map[string] string, len(ret))
	for _, v := range ret {
		suffix := ".u.qiniudn.com"
		bucketDomain[v.Bucket + suffix] = v.Domain
	}
	for k, bucket := range buckets {
		suffix := ".u.qiniudn.com"
		r[k] = Txws {
			Qiniudn: bucket + suffix,
		}
		if domain, ok := bucketDomain[bucket + suffix]; ok {
			r[k].Domain = domain
		}
	}
	c.ReplyIfError(err)
	c.ReplyObj(r)
}
