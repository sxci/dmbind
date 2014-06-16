package main

// run: source /Users/cheney/Projects/qiniu/env.sh && source /Users/cheney/Projects/qiniu/dmbind/env.sh && rm -r pkg; go run run.go

import (
	"encoding/json"
	"net/http"
	"os"
)

import (
	"dmbind/admin"
	"dmbind/api"
	"dmbind/cnc"
	"dmbind/dnspod"
	"dmbind/domain"
	"dmbind/lib/context"
	"dmbind/lib/pub"
	"dmbind/log"
	"dmbind/mail"
	"dmbind/stat"
	"dmbind/test"
)

var (
	bindHost = ":8088"
	confFile = "qboxdmbind.conf"
)

type Conf struct {
	Port   string         `json:"port"`
	Cnc    CncConf        `json:"cnc"`
	Pub    PubConf        `json:"pub"`
	Mail   MailConf       `json:"mail"`
	Admin  AdminConf      `json:"admin"`
	Dnspod dnspod.Setting `json:"dnspod"`
}

type CncConf struct {
	User string `json:"user"`
	Pswd string `json:"pswd"`
}

type AdminConf struct {
	User                string   `json:"user"`
	Pswd                string   `json:"pswd"`
	CacheTime           int64    `json:"cache_time"`
	HostName            string   `json:"host_name"`
	BindBucket          []string `json:"bind_bucket"`
	CdnSourceUrl        []string `json:"cdn_source_url"`
	IcpNotifyMail       []string `json:"icp_notify_mail"`
	TestFailNotifyMails []string `json:"test_fail_notify_mails"`
}

type PubConf struct {
	AccessKey string `json:"access_key"`
	SecretKey string `json:"secret_key"`
}

type MailConf struct {
	Profile           map[string]mail.MailProfile `json:"profile"`
	MailServerProfile string                      `json:"mail_server_profile"`
}

func init() {
	f, err := os.Open(confFile)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	var conf Conf
	err = json.NewDecoder(f).Decode(&conf)
	if err != nil {
		panic(err)
	}
	bindHost = ":" + conf.Port
	cnc.Setup(conf.Cnc.User, conf.Cnc.Pswd)
	admin.Setup(conf.Admin.User, conf.Admin.Pswd, conf.Admin.CacheTime)
	admin.SetupIcpNotifyMail(conf.Admin.IcpNotifyMail)
	admin.SetupLocalhost(conf.Admin.HostName + bindHost)
	admin.SetupTest(conf.Admin.TestFailNotifyMails, conf.Admin.BindBucket, conf.Admin.CdnSourceUrl)
	pub.Setup(&pub.Mac{
		AccessKey: conf.Pub.AccessKey,
		SecretKey: []byte(conf.Pub.SecretKey),
	})
	mail.SetupProfile(conf.Mail.Profile)
	err = mail.SetupMailServerProfile("qiniu_pop")
	if err != nil {
		panic(err)
	}
	dnspod.Setup(conf.Dnspod)
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/checkAndPass", context.Wrap(api.CheckAndPass))
	mux.HandleFunc("/api/test/check", context.Wrap(api.TestCheckDomain))
	mux.HandleFunc("/api/test/bench", context.Wrap(api.TestBenchDomain))
	mux.HandleFunc("/api/icp", context.Wrap(api.Icp))
	mux.HandleFunc("/api/admin/available", context.Wrap(api.AdminAvailableDomain))
	mux.HandleFunc("/api/admin/pass", context.Wrap(api.AdminPassDomain))
	mux.HandleFunc("/api/submitAvailable", context.Wrap(api.SubmitAvailable))
	mux.HandleFunc("/api/addOne", context.Wrap(api.AddOne))
	mux.HandleFunc("/api/add", context.Wrap(api.Add))
	mux.HandleFunc("/api/cron/checkAvailable", context.Wrap(api.CronCheckAvailable))
	mux.HandleFunc("/api/cron/submit", context.Wrap(api.CronSubmitToCnc))

	mux.HandleFunc("/cnc/", context.Wrap(cnc.Index))
	mux.HandleFunc("/cnc/cancel", context.Wrap(cnc.Cancel))
	mux.HandleFunc("/cnc/deployPass", context.Wrap(cnc.DeployPass))
	mux.HandleFunc("/cnc/submit", context.Wrap(cnc.Submit))
	mux.HandleFunc("/cnc/list", context.Wrap(cnc.List))
	mux.HandleFunc("/cnc/add", context.Wrap(cnc.Add))
	mux.HandleFunc("/cnc/addSafe", context.Wrap(cnc.AddSafe))
	mux.HandleFunc("/cnc/testDeploy", context.Wrap(cnc.TestDeploy))
	mux.HandleFunc("/cnc/passCheck", context.Wrap(cnc.PassCheck))
	mux.HandleFunc("/cnc/deploy", context.Wrap(cnc.Deploy))
	mux.HandleFunc("/cnc/ticket", context.Wrap(cnc.Ticket))
	mux.HandleFunc("/cnc/ticketDetail", context.Wrap(cnc.TicketDetail))
	mux.HandleFunc("/cnc/waitFor", context.Wrap(cnc.WaitFor))
	mux.HandleFunc("/cnc/waitForFinish", context.Wrap(cnc.WaitForFinish))

	mux.HandleFunc("/test/curl", context.Wrap(test.Curl))
	mux.HandleFunc("/test/bench", context.Wrap(test.Alibench))

	mux.HandleFunc("/mail/send", context.Wrap(mail.Send))
	mux.HandleFunc("/mail/cnc", context.Wrap(mail.GetCncTestUrl))

	mux.HandleFunc("/domain/icp", context.Wrap(domain.Icp))
	mux.HandleFunc("/domain/realIcp", context.Wrap(domain.RealIcp))
	mux.HandleFunc("/domain/aliTest", context.Wrap(domain.AliTest))
	mux.HandleFunc("/domain/pub", context.Wrap(domain.Pub))
	mux.HandleFunc("/domain/unpub", context.Wrap(domain.Unpub))
	mux.HandleFunc("/domain/txws", context.Wrap(domain.TxWs))

	mux.HandleFunc("/admin/icp", context.Wrap(admin.AdminIcp))
	mux.HandleFunc("/admin/list", context.Wrap(admin.List))
	mux.HandleFunc("/admin/pass", context.Wrap(admin.Pass))
	mux.HandleFunc("/admin/deny", context.Wrap(admin.Deny))
	mux.HandleFunc("/admin/status", context.Wrap(admin.Status))
	mux.HandleFunc("/admin/report", context.Wrap(admin.Report))
	mux.HandleFunc("/admin/batchTestDomain", context.Wrap(admin.BatchTestDomains))
	mux.HandleFunc("/admin/testDomain", context.Wrap(admin.TestDomain))
	mux.HandleFunc("/admin/testCnc", context.Wrap(admin.TestCnc))
	mux.HandleFunc("/admin/filter", context.Wrap(admin.Filter))
	mux.HandleFunc("/admin/filterSubmit", context.Wrap(admin.FilterSubmit))
	mux.HandleFunc("/admin/autotest", context.Wrap(admin.AutoTest))

	mux.HandleFunc("/stat/txws", context.Wrap(stat.TxWs))
	mux.HandleFunc("/stat/undeal", context.Wrap(stat.Undeal))
	mux.HandleFunc("/stat/weekReport", context.Wrap(stat.WeekReport))

	// mux.HandleFunc("/dns/txcdn", context.Wrap(dnspod.CdnToTx))
	// mux.HandleFunc("/dns/add", context.Wrap(dnspod.DnspodCName))

	mux.HandleFunc("/log/view", context.Wrap(log.View))

	panic(http.ListenAndServe(bindHost, mux))
}
