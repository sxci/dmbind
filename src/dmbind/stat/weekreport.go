package stat

import (
	"strconv"
	"net/url"
	"dmbind/admin"
	"dmbind/mail"
	"dmbind/lib/context"
)

func WeekReport(c *context.Context) {
	mailTo := c.Strings("to", nil)
	var rets []admin.DomainInfo
	c.ReplyIfError(context.Call(&rets, admin.List, url.Values {
		"before": {strconv.Itoa(86400*7)},
		"uniqueField": {"host"},
		"status": {"2"},
	}))
	var denies []admin.DomainInfo
	c.ReplyIfError(context.Call(&denies, admin.List, url.Values {
		"before": {strconv.Itoa(86400*7)},
		"uniqueField": {"host"},
		"status": {"0"},
	}))
	html := `<html>`
	th_css := `"heigth:50px;padding:20px;background:#ddd"`
	html += `<table cellspacing="0"><tr><th style=`+th_css+`>域名</th><th style=`+th_css+`>cdn</th><th style="background:#ddd">备案号</th>`
	css := "padding:10px;border-bottom:1px solid #f0f0f0;background:#dff0d8"
	for _, ret := range rets {
		html += `<tr><td style="`+css+`">`+ret.Host+`</td><td style="`+css+`">`+cdnName(ret.Cdn)+`</td><td style="`+css+`">`+ret.Icp+"</td></tr>"
	}
	css = "padding:10px;border-bottom:1px solid #f0f0f0;background:#f2dede"
	for _, ret := range denies {
		html += `<tr><td style="`+css+`">`+ret.Host+`</td><td style="`+css+`">`+cdnName(ret.Cdn)+`</td><td style="`+css+`">`+ret.Icp+"</td></tr>"
	}
	html += "</table>"
	html += `</html>`
	if len(mailTo) > 0 {
		c.ReplyIfError(context.Call(nil, mail.Send, url.Values {
			"to": mailTo,
			"subject": {"一周域名审核总揽"},
			"content": {html},
		}))
	}
	c.ReplyHtml(html)
}
