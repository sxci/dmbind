package cnc

import (
	"bytes"
	"dmbind/lib/fetcher"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var cncFetcher = fetcher.NewFetcher("myconf.chinanetcenter.com")
var (
	regexLoginLt           = regexp.MustCompile(`name="lt"[^"]+"([^"]+)`)
	regexErrInfo           = regexp.MustCompile(`font style=[^>]+>([^<]+)`)
	regexCheckLoginSuccess = regexp.MustCompile(`location.href="([^"]+)`)
	reJsRedirect           = regexp.MustCompile(`location.href="([^"]+)`)
	defDomainSrc           = "ws1.source.qbox.me"
)
var customerId = "51429"

var HEART_BEAT_DURATION = 60 * time.Second
var fetcherRestoreFile = "fetcher.tmp"

var (
	notifyEmail = "domain@qiniu.com;chenye@qiniu.com;xuzhaokui@qiniu.com;hantuo@qiniu.com"
)

func init() {
	//  cncFetcher.AutoHost = true
	// err := restoreFetcher()
	// log.Println("cnc.init: restore fetcher", err == nil, err)
	// if checkLogined() {
	// 	c, err := get("https://portal.chinanetcenter.com/smm/logo.action")
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	log.Println(string(c))
	// 	go keepAlive()
	// }
}

func checkLogined() (ok bool) {
	body, err := get("https://portal.chinanetcenter.com/smm/logo.action")
	if err != nil {
		return false
	}
	if strings.Contains(string(body), "七牛云存储") {
		return true
	}
	return
}

func restoreFetcher() (err error) {
	restore, err := ioutil.ReadFile(fetcherRestoreFile)
	if err != nil {
		return
	}
	cncFetcher, err = fetcher.Restore(string(restore))
	return
}

func storeFetcher() (err error) {
	s, err := cncFetcher.Store()
	if err != nil {
		return
	}
	err = ioutil.WriteFile(fetcherRestoreFile, []byte(s), os.FileMode(0666))
	return
}

func getLoginVerify() (imgBase64, lt string, err error) {
	_, body, err := cncFetcher.GetWithNoCache("http://portal.chinanetcenter.com/cas/login?request_locale=zh_CN")
	if err != nil {
		println("err1: ", err.Error())
		return
	}
	imgBase64, err = cncFetcher.GetBase64("http://portal.chinanetcenter.com/cas/captchaImage")
	if err != nil {
		println("err2: ", err.Error())
		return
	}

	println("body: ", string(body))

	lts := regexLoginLt.FindSubmatch(body)
	for i, v := range lts {
		println("lts: ", i, string(v))
	}
	if len(lts) < 2 {
		redirectInfo := regexCheckLoginSuccess.FindSubmatch(body)
		if len(redirectInfo) > 1 {
			_, body, err = cncFetcher.Get(string(redirectInfo[1]))
			if err != nil {
				return
			}
			println(body)
			return
		}
		err = errors.New("can't match lt")
		return
	}
	lt = string(lts[1])
	return
}

func login(user, pswd, verify, lt string) (err error) {
	u := "http://portal.chinanetcenter.com/cas/login?request_locale=zh_CN&service="
	u += "https%253A%252F%252Fportal.chinanetcenter.com%252Fsmm%252Fnews.action"
	_, body, err := cncFetcher.PostForm(u, url.Values{
		"username": {user},
		"password": {pswd},
		"jcaptcha": {verify},
		"submit":   {"登录"},
		"lt":       {lt},
	})
	errInfo := regexErrInfo.FindSubmatch(body)
	if len(errInfo) >= 2 {
		err = errors.New(string(errInfo[1]))
		return
	}
	redirectInfo := regexCheckLoginSuccess.FindSubmatch(body)
	if len(redirectInfo) <= 1 {
		err = errors.New("cnc.login: failure, redirect url not found!")
	}
	_, body, err = cncFetcher.Get(string(redirectInfo[1]))
	if err != nil {
		return
	}
	println(string(body))
	return
}

func get(url string) (body []byte, err error) {
	_, body, err = cncFetcher.Get(url)
	if err != nil {
		return
	}
	jsRedir := reJsRedirect.FindSubmatch(body)
	if len(jsRedir) <= 1 {
		return
	}
	_, body, err = cncFetcher.Get(string(jsRedir[1]))
	return
}

func post(path string, params url.Values) (body []byte, err error) {
	_, body, err = cncFetcher.PostForm(path, params)
	if err != nil {
		return
	}
	jsRedir := reJsRedirect.FindSubmatch(body)
	if len(jsRedir) <= 1 {
		return
	}
	_, body, err = cncFetcher.Get(string(jsRedir[1]))
	return
}

func call(ret interface{}, path string, params url.Values) (err error) {
	body, err := post(path, params)
	if err != nil {
		return
	}
	err = json.Unmarshal(body, ret)
	if err != nil {
		log.Println("cnc.call: result not json", string(body))
	}
	return
}

var alreadyKeepAlive = false

func keepAlive() (err error) {
	if alreadyKeepAlive {
		return
	}
	alreadyKeepAlive = true
	defer func() { alreadyKeepAlive = false }()

	data := [][]byte{
		[]byte("认证中心"),
		[]byte("Authentication Center"),
	}
	for {
		log.Println("send heart beat...")
		_, body, _ := cncFetcher.Get("http://portal.chinanetcenter.com/smm/logo.action")

		if bytes.Index(body, data[0])+bytes.Index(body, data[1]) > 0 {
			err = errors.New("bad auth")
			log.Println("cnc session invalid, need relogin!!")
			return
		}
		err = storeFetcher()
		if err != nil {
			log.Println("store fetcher", err == nil, err)
			err = nil
		}
		time.Sleep(HEART_BEAT_DURATION)
	}
	return
}

type DomainId struct {
	Domain string `json:"domainName"`
	Id     int    `json:"domainId"`
}

type RecordInfo struct {
	CustSuitType      string `json:"custSuitType"`
	CustomerDetailId  int    `json:"customerDetailId"`
	CustomerName      string `json:"customerName"`
	DeployTime        string `json:"deployTime"`
	Emails            string `json:"emails"`
	EndTime           string `json:"endTime"`
	RelevanceDomain   string `json:"relevanceDomain"`
	RelevanceDomainId int    `json:"relevanceDomainId"`
	Remark            string `json:"remark"`
	RequirementId     string `json:"requirementId"`
	StartTime         string `json:"startTime"`
	Status            int    `json:"status"`
	SubmitId          int    `json:"submitId"`
	SubmitName        string `json:"submitName"`
	TaskVersion       string `json:"taskVersion"`
	Version           string `json:"version"`
}

func versionList(version, status string) (infos []RecordInfo, err error) {
	body, err := post("/wsPortal/conf/customer-domain!initCustomerDomainGrid.action", url.Values{
		"_gt_json":         {versionListForm},
		"q_status_eq_i":    {status},
		"q_version_like_s": {version},
	})
	if err != nil {
		return
	}
	idx := bytes.Index(body, []byte("data : ["))
	if idx < 0 {
		err = errors.New("result not excepted! detail:" + string(body))
		return
	}
	err = json.Unmarshal(body[idx+7:len(body)-2], &infos)
	return
}

type VersioniDetail struct {
	CustomerDomainDetailId int    `json:"customerDomainDetailId"`
	CustomerVersion        string `json:"customerVersion"`
	TestUrl                string `json:"testUrl"`
	Icp                    string `json:"icp"`
	DomainName             string `json:"domainName"`
	ClientIp               string `json:"clientIp"`
	SrcIp                  string `json:"srcIp"`
}

func versionDetail(version string) (infos []VersioniDetail, err error) {
	body, err := post("/wsPortal/conf/customer-domain!initCustomerDomainDetailGrid.action", url.Values{
		"_gt_json":               {versionDetailForm},
		"q_customerVersion_eq_s": {version},
	})
	if err != nil {
		return
	}
	idx := bytes.Index(body, []byte("data : ["))
	if idx < 0 {
		err = errors.New("result not excepted! detail:" + string(body))
		return
	}
	err = json.Unmarshal(body[idx+7:len(body)-2], &infos)
	return
}

func list(match string, page, total int) (domains []DomainId, err error) {
	start := ((page - 1) * total) + 1
	end := start + total - 1
	body, err := post("/wsPortal/conf/customer-domain!getConfigFormDomains.action", url.Values{
		"_gt_json":                {fmt.Sprintf(listQuery, total, page, start, end)},
		"q_custSuitType_eq_s":     {"web"},
		"q_customerDetailId_eq_l": {customerId},
		"q_domainName_like_s":     {match},
	})
	idx := strings.Index(string(body), "data :")
	if idx < 0 {
		err = errors.New("cnc.list: result not match, " + string(body))
		return
	}
	err = json.Unmarshal(body[idx+6:len(body)-1], &domains)
	return
}

type InsertRecord struct {
	CustomerDomainDetailId string `json:"customerDomainDetailId"`
	CustomerVersion        string `json:"customerVersion"`
	DomainName             string `json:"domainName"`
	SrcIp                  string `json:"srcIp"`
	ClientIp               string `json:"clientIp"`
	TestUrl                string `json:"testUrl"`
	GtSn                   int    `json:"__gt_sn__"`
	GtNoValid              bool   `json:"__gt_no_valid__"`
	GtRowKey               string `json:"__gt_row_key__"`
}

// customerDomainDetailId":6563,"customerVersion":"20131231134917","testUrl":null,"icp":null,"domainName":"chenye.org","clientIp":"X-Forwarded-For","srcIp":"ws1.source.qbox.me","__gt_sn__":0,"__gt_row_key__":"__gt_addDomainGrid_r_0

func NewInsertRecord(idx int, domain, srcIp string) *InsertRecord {
	if srcIp == "" {
		srcIp = defDomainSrc
	}
	return &InsertRecord{
		DomainName: domain,
		ClientIp:   "X-Forwarded-For",
		SrcIp:      srcIp,
		GtNoValid:  true,
		GtSn:       idx,
		GtRowKey:   "__gt_addDomainGrid_r_" + strconv.Itoa(idx),
	}
}

type RetSave struct {
	Success   bool
	Exception string
	Msg       string
	Param     string
}

type RetSaveParams struct {
	IsCustomer bool   `json:"isCustomer"`
	Status     int    `json:"status"`
	Version    string `json:"version"`
}

func addDomain(records []*InsertRecord, relativeDomainId, action, version string) (ret RetSaveParams, err error) {
	data, err := json.Marshal(records)
	if err != nil {
		return
	}
	var retSave RetSave
	params := url.Values{
		"confCustomerDomain.emails":            {notifyEmail},
		"confCustomerDomain.custSuitType":      {"web"},
		"confCustomerDomain.customerDetailId":  {customerId},
		"domainJson":                           {""},
		"confCustomerDomain.status":            {"1"},
		"confCustomerDomain.relevanceDomainId": {relativeDomainId},
	}
	if version != "" {
		params["confCustomerDomain.version"] = []string{version}
	}
	domainJson := ""
	switch action {
	case "delete":
		domainJson = `{"insertedRecords":[],"updatedRecords":[],"deletedRecords":` + string(data) + `}`
	case "update":
		domainJson = `{"insertedRecords":[],"updatedRecords":` + string(data) + `,"deletedRecords":[]}`
	default:
		domainJson = `{"insertedRecords":` + string(data) + `,"updatedRecords":[],"deletedRecords":[]}`
	}
	params["domainJson"][0] = domainJson
	log.Println(action, params)

	err = call(&retSave, "/wsPortal/conf/customer-domain!save.action", params)
	if err != nil {
		return
	}
	if !retSave.Success {
		err = errors.New(retSave.Exception)
		return
	}
	err = json.Unmarshal([]byte(retSave.Param), &ret)
	return
}

func cancelTicket(version string) (ret RetSave, err error) {
	err = call(&ret, "/wsPortal/conf/customer-domain!operCustomerDomain.action", url.Values{
		"confCustomerDomain.version": {version},
		"confCustomerDomain.status":  {"7"},
	})
	if !ret.Success {
		err = errors.New(ret.Exception)
	}
	return

}

func passCheck(version string) (ret RetSave, err error) {
	err = call(&ret, "/wsPortal/conf/customer-domain!operCustomerDomain.action", url.Values{
		"confCustomerDomain.version":             {version},
		"confCustomerDomain.customerDetailId":    {customerId},
		"confCustomerDomain.status":              {"2"},
		"confCustomerDomain.isTestDeploySuccess": {"0"},
		"confCustomerDomain.remark":              {""},
		"isHaveReqCode":                          {"0"},
	})
	if !ret.Success {
		err = errors.New(ret.Exception)
	}
	return
}

func testDeploy(version string) (ret RetSave, err error) {
	err = call(&ret, "/wsPortal/conf/customer-domain!toTestDeploy.action", url.Values{
		"confCustomerDomain.version": {version},
	})
	if !ret.Success {
		err = errors.New(ret.Exception)
		return
	}
	testVer := ret.Param[1 : len(ret.Param)-1]
	tryTime := time.Now().Unix()
	tryTimes := 0
toVerifyResult:
	err = call(&ret, "/wsPortal/conf/customer-domain!getTestTaskStatus.action", url.Values{
		"testVersion":                {testVer},
		"confCustomerDomain.version": {version},
	})
	if !ret.Success {
		err = errors.New(ret.Exception)
		return
	}
	if ret.Param != "unfinished" {
		log.Println("cnc.testDeploy", "version:", version, "test success, tryTime:", tryTimes)
		return
	}
	if time.Now().Unix()-tryTime > 300 {
		err = errors.New("check deploy status too long(>180s), version " + version)
		return
	}
	time.Sleep(5 * time.Second)
	tryTimes++
	goto toVerifyResult
}

func deploy(version string) (ret RetSave, err error) {
	err = call(&ret, "/wsPortal/conf/customer-domain!toDeploy.action", url.Values{
		"confCustomerDomain.version": {version},
	})
	return
}

func submit(domains []string, source string) (err error) {
	if source == "" {
		source = "ws1.source.qbox.me"
	}
	domainNum := strconv.Itoa(len(domains) * 2)
	groupNum := strconv.Itoa(len(domains))
	content := "共" + groupNum + "组，" + domainNum + "个域名，分别是：\n\n"
	for i, domain := range domains {
		if i != 0 {
			content += ","
		}
		content += domain + ",." + domain
	}
	content += "\n源" + source + "\n配置单和.qiniudn.com一样"
	u := "http://resolve.chinanetcenter.com/cache.php?lan=zh-cn"
	mails := "hantuo1984@gmail.com;cynicholas@gmail.com;me@chenye.org;domain@qiniu.com;wangjinlei4ustc@gmail.com;liumm@chinanetcenter.com"
	data := url.Values{
		"caseCust":  {"七牛云存储"},
		"email":     {mails},
		"domain":    {"[多个（共" + domainNum + "个）]"},
		"sIP":       {source},
		"cbPicture": {"on"},
		"cbPage":    {"on"},
		"cbMedia":   {"on"},
		"cbOther":   {"on"},
		"cacheCB": {
			"JPG", "PNG", "JPEG", "GIF", "ICO", "BMP",
			"HTML", "HTM", "SHTML",
			"MP3", "WMA", "FLV", "MP4", "TS", "BAT", "WMV",
			"ZIP", "EXE", "RAR", "CSS", "JS", "TXT", "SWF",
		},
		"ustFile":      {""},
		"cacheTime":    {"1"},
		"selectTime":   {"天"},
		"sIPurl":       {""},
		"caseContents": {content},
		"submit":       {""},
		"timezone":     {""},
	}
	_, body, err := cncFetcher.PostForm(u, data)
	return
	if err != nil {
		return
	}
	ret := string(body)
	if !strings.Contains(ret, "message success") {
		log.Println("submit to cnc fail:", ret)
		err = errors.New("submit to cnc fail:" + ret)
	} else {
		log.Println("submit to cnc success,", domains)
	}
	return
}
