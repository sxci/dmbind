package dnspod

import (
	"errors"
	"dmbind/lib/fetcher"
	"dmbind/lib/context"
	"net/url"
)

type Setting struct {
	DomainId string `json:"domain_id"`
	Username string `json:"username"`
	Password string `json:"password"`
}
var setting Setting
var ErrExist = errors.New("record is exists!")

func Setup(set Setting) {
	setting = set
}

type DomainRecord struct {
	Status struct {
		Code string
		Message string
	}
	Record struct {
		Id string
		Name string
		Status string
	}
}

type RemarkRet struct {
	Status struct {
		Code string
		Message string
	}
}

func _CdnToTx(c *context.Context) {
	domains := c.MustStrings("domain")
	
	f := fetcher.NewFetcherHttps("dnsapi.cn")
	f.Header.Agent = "dnspod-python/0.01 (domain@qiniu.com; DNSPod.CN API v2.8)"
	var e error

	var ret DomainRecord
	for _, domain := range domains {
		err := f.CallPostForm(&ret, "/Record.Create", url.Values {
			"login_email": {setting.Username},
			"login_password": {setting.Password},
			"domain_id": {setting.DomainId},
			"sub_domain": {domain + ".qncdn.qiniudn.com"},
			"record_type": {"CNAME"},
			"record_line": {"默认"},
			"value": {"region.qiniudn.com.txcdn.cn"},
			"format": {"json"},
		})
		if err == nil && ret.Status.Code != "1" {
			if ret.Status.Message == "Subdomain roll record is limited" {
				err = ErrExist
			} else {
				err = errors.New(ret.Status.Message)
			}
		}
		if err != nil {
			e = err
			c.Error(err)
			continue
		}
		var mark RemarkRet
		err = f.CallPostForm(&mark, "/Record.Remark", url.Values {
			"login_email": {setting.Username},
			"login_password": {setting.Password},
			"domain_id": {setting.DomainId},
			"record_id": {ret.Record.Id},
			"remark": {"qncdn to tx"},
			"format": {"json"},
		})
		if err != nil { c.Error(err) }
		c.Info("success add dnspod record [", domain + "qncdn.qiniudn.com", ret.Record, "]")
	}
	c.ReplyIfError(e)
	c.ReplyObj(true)
}

func _DnspodCName(c *context.Context) {
	buckets := c.Strings("bucket", nil)
	if len(buckets) == 0 {
		c.ReplyErrorInfo("miss bucket")
		return
	}
	raise := c.Bool("raiseIfOne", false)

	f := fetcher.NewFetcherHttps("dnsapi.cn")
	f.Header.Agent = "dnspod-python/0.01 (domain@qiniu.com; DNSPod.CN API v2.8)"

	var e error
	for _, bucket := range buckets {
		var ret DomainRecord
		err := f.CallPostForm(&ret, "/Record.Create", url.Values {
			"login_email": {setting.Username},
			"login_password": {setting.Password},
			"domain_id": {setting.DomainId},
			"sub_domain": {bucket + ".u"},
			"record_type": {"CNAME"},
			"record_line": {"默认"},
			"value": {"d.qiniudn.com"},
			"format": {"json"},
		})
		if err == nil && ret.Status.Code != "1" {
			if ret.Status.Message == "Subdomain roll record is limited" {
				err = ErrExist
			} else {
				err = errors.New(ret.Status.Message)
			}
		}
		if raise { c.ReplyIfError(err) }
		if err != nil {
			e = err
			c.Error(err)
			continue
		}
		var mark RemarkRet
		err = f.CallPostForm(&mark, "/Record.Remark", url.Values {
			"login_email": {setting.Username},
			"login_password": {setting.Password},
			"domain_id": {setting.DomainId},
			"record_id": {ret.Record.Id},
			"remark": {"同兴暂时切网宿"},
			"format": {"json"},
		})
		if err != nil { c.Error(err) }
		c.Info("success add dnspod record [", bucket, ret.Record, "]")
	}
	c.ReplyIfError(e)
	c.ReplyObj(true)
}
