package domain

import (
	"regexp"
	"errors"
	"net/url"
	"strings"

	"dmbind/lib/context"
	"dmbind/lib/fetcher"
)

func RealIcp(c *context.Context) {
	domain := c.MustString("domain")
	restore, verify := c.String("token", ""), c.String("verify", "")
	if restore != "" && verify != "" {
		f, err := fetcher.Restore(restore)
		c.ReplyIfError(err)
		_, body, err := f.PostForm("/beian.aspx", url.Values {
			"s": {domain},
			"code": {verify},
		})
		c.ReplyIfError(err)
		d := string(body)
		if strings.Index(d, "许可证号") > 0 {
			c.ReplyObj(true)
		}
		if strings.Index(d, "未备案") > 0 {
			c.ReplyObj(false)
		}
		if strings.Index(d, "验证码错误") > 0 {
			c.ReplyErrorInfo("invalidCode")
		}
		c.ReplyErrorInfo("unexcept error")
	}

	f := fetcher.NewFetcher("tool.chinaz.com")
	img, err := f.GetBase64("/beian.aspx?at=img")
	c.ReplyIfError(err)
	store, err := f.Store()
	c.ReplyIfError(err)
	c.ReplyObj([]string{img, store})
}

func Icp(c *context.Context) {
	domains := c.Strings("domain", nil)
	multi := c.Bool("multi", false)
	selected := c.Bool("select", false)
	ret, length := make([][2]interface{}, len(domains)), 0
	for _, domain := range domains {
		yes, err := CheckRecordChinaZ(domain)
		if err != nil {
			c.ErrorWithTag("chinaz", err)
			err = nil
		}
		if yes {
			c.Info("check domain " + domain + " icp success, chinaz")
			ret[length] = [2]interface{} {domain, 1}
			length ++
			continue
		}
		yes, err = CheckRecordBeianbeian(domain)
		if err != nil {
			c.ErrorWithTag("beianbeian", err, domain)
			ret[length] = [2]interface{} {domain, -1}
			length ++
			continue
		}
		result := "failure"
		if yes { result = "success, beianbeian" }
		c.Info("check domain", domain, "icp", result)
		r := 1
		if ! yes { r = 0 }
		ret[length] = [2]interface{} {domain, r}
		length ++
	}
	ret = ret[:length]
	if len(domains) == 1 && ! multi && ! selected {
		if len(ret) <= 0 {
			c.ReplyErrorInfo("unknown fail")
		}
		code := ret[0][1].(int)
		if code == -1 {
			c.ReplyErrorInfo("unknown fail")
		}
		c.ReplyObj(code == 1)
	}
	if selected {
		data := make(map[string] []string, 2)
		data["icp"], data["unknown"] = make([]string, len(ret)), make([]string, len(ret))
		lengthIcp, lengthUnknown := 0, 0
		for _, i := range ret {
			switch i[1].(int) {
			case 1:				
				data["icp"][lengthIcp] = i[0].(string)
				lengthIcp++
			default:
				data["unknown"][lengthUnknown] = i[0].(string)
				lengthUnknown++
			}
		}
		data["icp"] = data["icp"][:lengthIcp]
		data["unknown"] = data["unknown"][:lengthUnknown]
		c.ReplyObj(data)
	}
	c.ReplyObj(ret)
}

func CheckRecordBeianbeian(domain string) (yes bool, err error) {
	tRe := regexp.MustCompile(`<table id="show_table" class`)
	iRe := regexp.MustCompile("/beianxinxi")
	host := "www.beianbeian.com"
	url := "/search/" + domain
	f := fetcher.NewFetcher(host)
	_, body, err := f.Get(url)
	if err != nil { return }
	if len(tRe.FindAll(body, -1)) == 0 {
		err = errors.New("server error")
		return
	}
	yes = len(iRe.FindAll(body, -1)) > 0
	return
}

func CheckRecordChinaZ(domain string) (yes bool, err error) {
	data := url.Values {
		"s": {domain},
	}
	f := fetcher.NewFetcher("tool.chinaz.com")
	_, body, err := f.PostForm("/beian.aspx", data)
	if err != nil { return }
	d := string(body)
	if strings.Index(d, "许可证号") > 0 {
		yes = true
		return
	}
	if strings.Index(d, "未备案") > 0 { return }
	err = errors.New("unexcept content")
	return
}
