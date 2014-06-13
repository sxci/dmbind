package admin

import (
	"time"
	"sync"
	"net/url"
	"net/http"
	"dmbind/mail"
	dm "dmbind/domain"
	"dmbind/lib/context"
)

type CncTestRet struct {
	Domains []string `json:"domains"`
	Url     string   `json:"url"`
}

func TestCnc(c *context.Context) {
	var rets []CncTestRet
	c.ReplyIfError(context.Call(&rets, mail.GetCncTestUrl, nil))
	retChan := make(chan dm.TestResult, len(rets))
	data := ""
	for _, ret := range rets {
		if len(ret.Domains) == 0 { continue }
		var wg sync.WaitGroup
		wg.Add(len(ret.Domains))
		for _, domain := range ret.Domains {
			go func(domain string) {
				var ret dm.TestResult
				err := context.Call(&ret, TestDomain, url.Values {
					"domain": {domain},
				})
				wg.Done()
				if err != nil {
					c.Error(err)
				}
				retChan <- ret
			}(domain)
		}
		wg.Wait()
		for ret := range retChan {
			if ret.Info == "" { continue }
			succ := "success: false\n"
			if ret.Success { succ = "success:true\n" }
			data += succ
			data += "testPath: " + ret.TestPath + "\n"
			data += ret.Info
			data += "\n\n"
		}
		_, err := http.Get(ret.Url)
		if err != nil { c.Error(err) }
	}
	if data != "" {
		err := reportError(data)
		if err != nil { c.Error(err) }
	}
	c.ReplyObj(rets)
}

func reportError(data string) (err error) {
	err = context.Call(nil, mail.Send, url.Values {
		"profile": {"noreply"},
		"to": TestFailNotifyMail,
		"subject": {time.Now().Format("2006-01-02 15:04:05") + " - 阿里测测试结果"},
		"content": {data},
	})
	return
}
