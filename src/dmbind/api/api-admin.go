package api

import (
	"strconv"
	"net/url"
	"dmbind/admin"
	"dmbind/lib/context"
)

func AdminAvailableDomain(c *context.Context) {
	var ret []string
	err := context.Call(&ret, admin.List, url.Values {
		"status": {"1"},
		"showList": {"1"},
		"field": {"host"},
	})
	c.ReplyIfError(err)
	var domains map[string] string
	err = context.Call(&domains, Icp, url.Values {"host": ret})
	c.ReplyIfError(err)
	c.ReplyObj(domains)
}

func AdminPassDomain(c *context.Context) {
	domain := c.MustString("domain")
	var ret []admin.DomainInfo
	err := context.Call(&ret, admin.List, url.Values {
		"status": {"3"},
		"domain": {domain},
		"field": {"id-uid-bucket-cdn-domain"},
	})
	c.ReplyIfError(err)
	for _, item := range ret {
		// if needDnspodRecname(item.Cdn) {
			// err := context.Call(nil, dnspod.DnspodCName, url.Values {
				// "bucket": {item.Bucket},
			// })
			// if err != nil && ! strings.Contains(err.Error(), "exist") {
				// c.Error(err, item)
				// continue
			// }
		// }
		err := context.Call(nil, admin.PassId, url.Values {
			"id": {item.Id},
			"domain": {item.Domain},
			"bucket": {item.Bucket},
			"uid": {strconv.Itoa(item.Uid)},
		})
		if err != nil {
			c.Error(err)
		}
	}
	context.Call(nil, admin.Report, url.Values {
		"domain": {domain},
		"mail": {"1"},
	})
	c.ReplyJSON(ret)
}

func needDnspodRecname(cdn int) bool {
	return cdn == 1
}
