package api

import (
	"strings"
	"dmbind/lib/context"
	"dmbind/admin"
	"dmbind/cnc"
	"dmbind/stat"
	"dmbind/mail"
	"encoding/json"
	"time"
	"net/url"
)

var Hostname = "183.60.175.39:8088"

func CronCheckAvailable(c *context.Context) {
	var hostList []string
	err := context.Call(&hostList, admin.List, url.Values {
		"status": {"1"},
		"field": {"host"},
		"showList": {"1"},
	})
	c.ReplyIfError(err)
	var icpList map[string] string
	err = context.Call(&icpList, Icp, url.Values {
		"domain": hostList,
	})
	c.ReplyIfError(err)
	submitDomains := make([]string, len(hostList))
	length := 0
	for _, host := range hostList {
		if icpList[host] == "" {
			// noicp
			continue
		}
		submitDomains[length] = host
		length ++
	}
	submitDomains = submitDomains[:length]
	
	var ret cnc.SafeAddRet
	err = context.Call(&ret, Add, url.Values {
		"domain": {},
	})
	c.ReplyObj(submitDomains)
}

func sendMail(content string) {
	context.Call(nil, mail.Send, url.Values {
		"profile": {"chenye_qiniu"},
		"to": {"me@chenye.org"},
		"subject": {"[dmbind] 数据提交给网宿cdn，" + time.Now().Format("2006-01-02")},
		"content": {content},
	})
}

func inArray(a string, b []string) bool {
	for _, c := range b { if c == a { return true }}
	return false
}

func CronSubmitToCnc(c *context.Context) {
	exclude := c.Strings("exclude", nil)
	needEmail := strings.HasPrefix(c.Req.Host, "localhost")
	var domains []string
	err := context.Call(&domains, stat.Undeal, url.Values {
		"json": {"1"},
		"before": {"0"},
	})
	if needEmail && err != nil{
		sendMail(err.Error())
		return
	} else {
		c.ReplyIfError(err)
	}
	movePos := 0
	for idx, d := range domains {
		if inArray(d, exclude) {
			movePos ++
			continue
		}
		domains[idx-movePos] = d
	}
	domains = domains[:len(domains)-movePos]

	tryTime := 0
retry:
	var ret cnc.SafeAddRet
	err = context.Call(&ret, Add, url.Values {
		"domain": domains,
	})
	if err != nil && strings.Contains(err.Error(), "系统发生异常") {
		tryTime ++
		if tryTime < 3 {
			time.Sleep(time.Second)
			goto retry
		}
	}
	if needEmail && err != nil {
		errInfo := strings.Replace(err.Error(), "<br>", "\n", -1)
		sendMail(errInfo)
		return
	} else {
		c.ReplyIfError(err)
	}
	output, _ := json.Marshal(ret)
	content := string(output) + "\n\n" + strings.Join(domains, "\n")
	content += "\n\nhttp://" + Hostname + "/cnc/deploy?version=" + ret.Ret.Version
	content += "\nhttp://" + Hostname + "/cnc/ticket?version=" + ret.Ret.Version
	content += "\nhttp://" + Hostname + "/cnc/ticketDetail?version=" + ret.Ret.Version
	sendMail(content)
	c.ReplyObj(ret)
}
