package admin

import (
	"net/url"
	"dmbind/domain"
	"dmbind/lib/context"
	// "dmbind/tx"
)

func AdminIcp(c *context.Context) {
	domains := c.MustStrings("domain")
	test := c.Bool("test", false)


	if len(domains) > 1 {
		html := `<html>`
		for _, domain := range domains {
			html += `<iframe border="0" style="border:none;width: 220px;height: 140px;" src="/admin/icp?domain=`+domain+`"></iframe>`
		}
		html += `</html>`
		c.ReplyHtml(html)
		return
	}

	dm := domains[0]

	force := c.Bool("force", false)
	if ! force {
		var domains []string
		err := context.Call(&domains, List, url.Values {
			"status": {"1"},
			"domain": {dm},
		})
		c.ReplyIfError(err)
		if len(domains) == 0 {
			c.ReplyErrorInfo("domain status not 1!")
		}
	}

	if ! c.IsPost() {
		var ret []string
		err := context.Call(&ret, domain.RealIcp, url.Values{"domain": {dm}})
		c.ReplyIfError(err)
		img, store := ret[0], ret[1]
		html := `<html>
		<style>
		#verify{background-image: url(data:image/png;base64,`+img+`);background-repeat: no-repeat;width:200px; height: 50px}
		input {width:200px; heigth:30px;padding: 5px;font-size: 14pt;}
		</style>
		<body>
		<div id="verify"></div>
		<form action="" method="POST">
		`+dm+`<br/>
		<input type="input" name="verify" />
		<input type="hidden" value="`+store+`" name="token" />
		</form>
		</body>
		</html>`
		c.ReplyHtml(html)
		return
	}

	var ret bool
	verify, token := c.MustString("verify"), c.MustString("token")
	err := context.Call(&ret, domain.RealIcp, url.Values {
		"verify": {verify},
		"token": {token},
		"domain": {dm},
	})
	if err != nil && err.Error() == "invalidCode" {
		path := "?domain="+dm
		if test { path += "&test=1" }
		c.Redirect(path)
		return
	}
	c.ReplyIfError(err)
	if test {
		c.ReplyObj(ret)
		return
	}
	c.Info("check icp of", dm, ",", ret)
	if ! ret {
		c.ReplyIfError(context.Call(nil, Deny, url.Values {
			"domain": {dm},
			"from": {"1"},
		}))
		c.ReplyObj(false)
		return
	}
	c.ReplyIfError(context.Call(nil, Status, url.Values{
		"domain": {dm},
		"from": {"1"},
		"status": {"3"},
	}))
	c.ReplyObj(true)
}

func InIntArray(a int, b []int) bool {
	for _, i := range b {
		if i == a { return true }
	}
	return false
}
