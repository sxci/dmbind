package stat

import (
	"time"
	"net/url"
	"strconv"
	"dmbind/admin"
	"dmbind/lib/context"
	"dmbind/mail"
)

var cdnMap = []string {"网宿", "同兴"}

func cdnName(code int) string {
	return cdnMap[code]
}

func Undeal(c *context.Context) {
	cdn := c.Strings("cdn", nil)
	before := c.String("before", "-259200")
	mailAddr := c.String("mail", "")
	profile := c.String("profile", "alert")
	domainOnly := c.Bool("domainOnly", false)
	api := c.Bool("api", false)
	useJson := c.Bool("json", false)
	var ret []admin.DomainInfo
	c.ReplyIfError(context.Call(&ret, admin.List, url.Values {
		"cdn": cdn,
		"status": {"3"},
		"field": {"host-update_time-cdn"},
		"uniqueField": {"host"},
		"before": {before},
	}))
	if len(ret) == 0 {
		c.ReplyText("")
	}
	text := ""
	t := time.Now().Unix()
	if api {
		for idx, domain := range ret {
			if idx > 0 { text += "&" }
			text += "domain=" + domain.Host
		}
		return
	} else if useJson {
		domains := make([]string, len(ret))
		for idx, data := range ret {
			domains[idx] = data.Host
		}
		c.ReplyObj(domains)
	} else {
		for _, i := range ret {
			text += i.Host
			if ! domainOnly {
				text += ", " + cdnName(i.Cdn) + ", " + strconv.Itoa(int(t-i.UpdateTime)/86400) + "day"
			}
			text += "\n"
		}
	}
	if mailAddr != "" {
		data, err := strconv.Atoi(before)
		if err != nil { return }
		t := time.Now().Add(time.Duration(data))
		c.ReplyIfError(context.Call(nil, mail.Send, url.Values {
			"to": {mailAddr},
			"profile": {profile},
			"subject": {"未处理自定义域名清单, 截止到" + t.Format("2006-01-02")},
			"content": {text},
		}))
	}
	c.ReplyText(text)
}
