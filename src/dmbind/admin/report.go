package admin

import (
	"fmt"
	"net/url"

	"dmbind/lib/context"
	"dmbind/mail"
)

func Report(c *context.Context) {
	domains := c.MustStrings("domain")
	before := c.String("before", "86400")
	sendMail := c.Bool("mail", false)
	html := ""
	for _, domain := range domains {
		var rets []DomainInfo
		c.ReplyIfError(context.Call(&rets, List, url.Values {
			"domain": {domain},
			"before": {before},
		}))
		h := reportTpl(rets)
		html += h
		if sendMail && len(rets) > 0 {
			c.ReplyIfError(context.Call(nil, mail.Send, url.Values {
				"to": {rets[0].Email},
				"subject": {"[七牛自定义域名] 审核进度通知"},
				"content": {h},
			}))
			c.ReplyIfError(context.Call(nil, mail.Send, url.Values {
				"to": {"domain@qiniu.com"},
				"subject": {"[七牛自定义域名] 审核进度通知"},
				"content": {h},
			}))
			c.Info("send report to " + rets[0].Email + ", domain:" + domain)
		}
	}
	c.ReplyHtml(html)
}

func reportTpl(rets []DomainInfo) (data string) {
	helpDesc := 0
	domainString := ""
	for _, ret := range rets {
		switch ret.Status {
		case 0:
			helpDesc |= 1
		case 1:
			helpDesc |= 2
		case 2:
			helpDesc |= 4
		case 3:
			helpDesc |= 8
		case 4:
			helpDesc |= 2
		}
		domainString += domainRetTpl(ret)
	}
	html := `
	<html><body>
	Hello, 欢迎使用七牛云存储自定义域名绑定服务。<br>
	您的自定义域名审核情况如下：
	<table cellspacing="0" style="border:0px;margin-top:20px"><tr>
	<th style="border-bottom: 1px solid #e5e5e5;padding:10px;min-width:100px;background:#eeeeee">bucket</th>
	<th style="border-bottom: 1px solid #e5e5e5;padding:10px;min-width:150px;background:#eeeeee">域名</th>
	<th style="border-bottom: 1px solid #e5e5e5;padding:10px;min-width:150px;background:#eeeeee">状态</th>
	<th style="border-bottom: 1px solid #e5e5e5;padding:10px;min-width:150px;background:#eeeeee">查看详情</th>
	</tr>`
	html += domainString
	html += `</table><div style="margin-top: 20px">`
	if helpDesc & 1 != 0 {
		html += `如果您的域名未备案，请在完成网站备案后，再次提交申请。<br />`
	}
	if helpDesc & 2 != 0 {
		html += `域名一般的审核时间是2-3天 <br />`
	}
	if helpDesc & 4 != 0 {
		html += `如果您的域名已通过申请，请到域名的DNS管理中配置CNAME(别名)后即可使用。<br /> `
		html += `帮助信息：<a href="http://support.qiniu.com/entries/24881791-%E4%B8%83%E7%89%9B%E4%BA%91%E5%AD%98%E5%82%A8%E8%87%AA%E5%AE%9A%E4%B9%89%E5%9F%9F%E5%90%8D%E7%BB%91%E5%AE%9A%E6%B5%81%E7%A8%8B">如何配置CNAME？</a><br />`
	}
	if helpDesc & 8 != 0 {
		html += `配置加速的过程一般需要1-2天 <br />`
	}

	html += `<br />
	<br />
	七牛云存储团队<br />
	<br />
	登录地址：<a href="https://portal.qiniu.com">https://portal.qiniu.com</a><br />
	客服电话：400-808-9176<br />
	客服邮箱：support@qiniu.com<br />
	在线支持：<a href="http://support.qiniu.com">https://support.qiniu.com</a><br />
	企业微博：<a href="http://weibo.com/qiniutek">http://weibo.com/qiniutek</a><br />
	<br />
	------<br />
	<br />
	温馨提示：此邮件由系统自动发送，请勿直接回复。
	</div></body></html>`
	return html
}

func domainRetTpl(ret DomainInfo) (data string) {
	status := "审核中"
	statusStyle := "border-bottom: 1px solid #e5e5e5;background-color:#ddd;text-align:center;padding:10px;"
	if ret.Status == 2 || ret.Status == 4 {
		statusStyle += "color:#468847; background-color:#dff0d8"
		status = "已通过<br />CNAME:" + ret.Domain + ".qncdn.qiniudn.com"
	} else if ret.Status == 0 {
		statusStyle += "color:#b94a48;background-color:#f2dede"
		status = "未备案"
	} else if ret.Status == 3 {
		status = "配置加速中"
	}
	tpl := `<tr>
	<td style="border-bottom: 1px solid #e5e5e5;padding:10px">
		<a href="https://portal.qiniu.com/bucket/setting/basic?bucket=%s">%s</a>
	</td>
	<td style="border-bottom: 1px solid #e5e5e5;padding:10px">%s</td>
	<td style="%s">%s</td>
	<td style="border-bottom: 1px solid #e5e5e5;padding:10px;text-align:center">
		<a href="https://portal.qiniu.com/bucket/setting/basic?bucket=%s">查看</a>
	</td></tr>`
	return fmt.Sprintf(tpl, ret.Bucket, ret.Bucket, ret.Domain, statusStyle, status, ret.Bucket)
}
