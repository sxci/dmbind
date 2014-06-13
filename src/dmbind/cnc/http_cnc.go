package cnc

import (
	"errors"
	"strconv"
	"time"
	"net/url"
	"strings"

	"dmbind/lib/context"
	"dmbind/mail"
)

var (
	cncUser string
	cncPswd string
)

func Setup(user, pswd string) {
	cncUser = user
	cncPswd = pswd
}

func List(c *context.Context) {
	page := c.Int("page", 1)
	total := c.Int("total", 30)
	search := c.String("search", "")
	
	ret, err := list(search, page, total)
	c.ReplyIfError(err)
	c.ReplyJSON(ret)
}

type SafeAddRet struct {
	Ret RetSaveParams `json:"ret"`
	Unicp []string `json:"unicp"`
}

func AddSafe(c *context.Context) {
	domains := c.MustStrings("domain")
	src := c.String("src", "")
	var ret RetSaveParams
	err := context.Call(&ret, Add, url.Values {
		"domain": domains,
		"src": {src},
	})
	c.ReplyIfError(err)

	details, err := versionDetail(ret.Version)
	c.ReplyIfError(err)

	var newDetail []*InsertRecord
	var unicp []string
	for i, detail := range details {
		if detail.Icp == "" {
			unicp = append(unicp, detail.DomainName)
			nd := NewInsertRecord(i, detail.DomainName, detail.SrcIp)
			nd.CustomerDomainDetailId = strconv.Itoa(detail.CustomerDomainDetailId)
			nd.CustomerVersion = detail.CustomerVersion
			newDetail = append(newDetail, nd)
		}
	}

	// !!
	records, err := versionList(ret.Version, "")
	if err == nil && len(records) == 0 {
		err = errors.New("result not found")
	}
	c.ReplyIfError(err)
	record := records[0]

	if len(unicp) > 0 {
		ret, err = addDomain(newDetail, strconv.Itoa(record.RelevanceDomainId), "delete", ret.Version)
		c.ReplyIfError(err)
	}
	c.ReplyObj(SafeAddRet{ret, unicp})
}

func Add(c *context.Context) {
	domains := c.MustStrings("domain")
	src := c.String("src", "")
	relId, ok := getIdleRelativeId()
	if ! ok {
		c.ReplyErrorInfo("not idle relativeDomain id")
		return
	}
	ds := make([]*InsertRecord, len(domains))
	for i, d := range domains {
		if d == "" { continue }
		ds[i] = NewInsertRecord(i, d, src)
	}
	ret, err := addDomain(ds, relId, "insert", "")
	if err != nil {
		finishRelativeId(relId)
		c.ReplyError(err)
	}
	fillVersion(relId, ret.Version, domains)
	c.Info("ticket #", ret.Version, "created! relId:", relId, "domain:", domains)
	c.ReplyJSON(ret)
}

func PassCheck(c *context.Context) {
	version := c.MustString("version")
	ret, err := passCheck(version)
	c.ReplyIfError(err)
	domains := getDomainsFromVersion(version)
	c.Info("ticket #", version, "passCheck! domains:", domains)
	c.ReplyJSON(ret)
}

func TestDeploy(c *context.Context) {
	version := c.MustString("version")
	ret, err := testDeploy(version)
	c.ReplyIfError(err)
	c.ReplyJSON(ret)
}

func TicketDetail(c *context.Context) {
	version := c.MustString("version")
	ret, err := versionDetail(version)
	c.ReplyIfError(err)
	c.ReplyJSON(ret)
}

func WaitFor(c *context.Context) {
	version := c.MustString("version")
	exceptStatus := c.MustString("status")
	for {
		time.Sleep(time.Minute)
		infos, err := versionList(version, exceptStatus)
		if err != nil { continue }
		if len(infos) != 1 { continue }
		c.ReplyObj(true)
	}
}

func WaitForFinish(c *context.Context) {
	version := c.MustString("version")
	err := context.Call(nil, WaitFor, url.Values {
		"version": {version},
		"status": {"5"},
	})
	c.ReplyIfError(err)
	domains := getDomainsFromVersion(version)
	c.Info("ticket #", version, "finish. domain:", domains)
	finishRelativeId(version)
	c.ReplyObj(true)
}

func Cancel(c *context.Context) {
	version := c.MustString("version")
	ret, err := cancelTicket(version)
	c.ReplyIfError(err)
	c.Info("cancel ticket #", version)
	c.ReplyObj(ret)
}

func Deploy(c *context.Context) {
	version := c.MustString("version")
	ret, err := deploy(version)
	c.ReplyIfError(err)
	domains := getDomainsFromVersion(version)
	c.Info("ticket #", version, " deploy, domains:", domains)
	c.ReplyJSON(ret)
}

func Ticket(c *context.Context) {
	status := c.String("status", "")
	version := c.String("version", "")
	ret, err := versionList(version, status)
	c.ReplyIfError(err)
	c.ReplyJSON(ret)
}

func DeployPass(c *context.Context) {
	ret, err := versionList("", "2")
	c.ReplyIfError(err)
	rets := make(map[string] interface{}, len(ret))
	for _, v := range ret {
		version := v.Version
		rets[version], err = deploy(version)
		if err != nil {
			rets[version] = err.Error()
		}
	}
	c.ReplyJSON(rets)
}

func Index(c *context.Context) {
	if ! c.IsPost() {
		img, lt, err := getLoginVerify()
		if err != nil {
			c.ReplyError(err)
			return
		}
		errInfo := c.String("error", "")
		verify := "#verify {background-image:url(data:image/png;base64,"
		verify += img + ");background-repeat:no-repeat;width:80px;height:20px;float:left}"
		html := `<html>
		<style>` + verify + `</style>
		<body>
		<div id="verify"></div>
		<form action="/cnc/" method="POST">
		<input type="hidden" name="lt" value="`+lt+`" />
		<input type="input" name="verify" autofocus/>
		<input type="submit" />
		</form>
		<label>` + errInfo + `</label>
		</body>
		</html>`
		c.ReplyHtml(html)
		return
	}
	lt := c.MustString("lt")
	verify := c.MustString("verify")
	err := login(cncUser, cncPswd, verify, lt)
	if err != nil {
		c.Redirect("?" + url.Values{"error": {err.Error()}}.Encode())
		return
	}
	_, err = get("http://myconf.chinanetcenter.com/wsPortal/conf/customer-domain-add.action")
	c.ReplyIfError(err)
	go keepAlive()
}

func Submit(c *context.Context) {
	domains := c.MustStrings("domain")
	source := c.String("source", "ws1.source.qbox.me")
	err := submit(domains, source)
	c.ReplyIfError(err)
	c.ReplyIfError(context.Call(nil, mail.Send, url.Values {
		"to": {"chenye@qiniu.com"},
		"subject": {"发送增加域名信息给cdn，请确认"},
		"content": {strings.Join(domains, "\n")},
	}))
	c.ReplyObj(true)
}
