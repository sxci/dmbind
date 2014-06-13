package admin

import (
	"log"
	"sync"
	"time"
	"bytes"
	"regexp"
	"errors"
	"strings"
	"net/url"
	"encoding/json"
)

import (
	"dmbind/domain"
	"dmbind/lib/context"
	"dmbind/lib/fetcher"
)

var (
	User string
	Pswd string
	adminFetcher = fetcher.NewFetcherHttps("admin.qiniu.com")
	CacheTime int64 = 10
)

type DomainInfo struct {
	Id         string `json:"id,omitempty"`
	Uid        int    `json:"uid,omitempty"`
	Cdn        int    `json:"cdn,omitempty"`
	Icp        string `json:"icp,omitempty"`
	Email      string `json:"email,omitempty"`
	Domain     string `json:"domain,omitempty"`
	Bucket     string `json:"bucket,omitempty"`
	Host       string `json:"host,omitempty"`
	Status     int    `json:"status,omitempty"`
	CreateTime int64  `json:"create_time,omitempty"`
	UpdateTime int64  `json:"update_time,omitempty"`
}

func init() {
	adminFetcher.CacheTime = CacheTime
}

func Setup(user, pswd string, cacheTime int64) {
	User, Pswd = user, pswd
	adminFetcher.CacheTime = cacheTime
}

func Info(o ...interface{}) {
	o = append([]interface{}{"[admin]"}, o...)
	log.Println(o...)
}

func getAdmin(path string, cache bool) (body []byte, err error) {
	f := adminFetcher.Get
	if ! cache { f = adminFetcher.GetWithNoCache }
	_, body, err = f(path)
	if err != nil { return }
	if ! isNeedLogin(body) { return }
	adminFetcher.RemoveGetCache(path)
	if err = login(User, Pswd); err != nil { return }
	_, body, err = adminFetcher.Get(path)
	return
}

func postAdmin(path string, params url.Values) (body []byte, err error) {
	resp, body, err := adminFetcher.PostForm(path, params)
	if err != nil { return }
	if ! isNeedLogin(body) {
		if resp.StatusCode / 100 != 2 {
			err = errors.New(string(body))
			body = nil
		}
		return
	}
	if err = login(User, Pswd); err != nil { return }
	resp, body, err = adminFetcher.PostForm(path, params)
	if resp.StatusCode / 100 != 2 {
		err = errors.New(string(body))
		body = nil
	}
	return
}

func callGetAdmin(ret interface{}, path string, cache bool) (err error) {
	body, err := postAdmin(path, nil)
	if err != nil { return }
	err = json.Unmarshal(body, ret)
	return
}

func callPostAdmin(ret interface{}, path string, params url.Values) (err error) {
	body, err := postAdmin(path, params)
	if err != nil { return }
	err = json.Unmarshal(body, ret)
	return
}

func isNeedLogin(body []byte) bool {
	return bytes.Contains(body, []byte(`id="login"`)) || bytes.Contains(body, []byte("CSRF")) || bytes.Contains(body, []byte("Error"))
}

var loginLock sync.Mutex
func login(user, pswd string) (err error) {
	loginLock.Lock()
	defer loginLock.Unlock()
	// check whether logined
	_, body, err := adminFetcher.GetWithNoCache("/domainmgr/")
	if err != nil { return }
	if ! isNeedLogin(body) { return }

	Info("try login...")
	params := url.Values {
		"username": {user},
		"password": {pswd},
	}
	_, body, err = adminFetcher.PostForm("/login", params)
	if err != nil {
		Info("login fail: " + err.Error())
		return
	}
	if body[0] == '{' && body[len(body)-1] == '}' {
		info := string(body)
		Info("login fail: " + info)
		err = errors.New(info)
		return
	}
	adminFetcher.GetWithNoCache("/domainmgr/?status=1")
	Info("login success")
	return
}

func mapToList(datas []map[string] interface{}, key string) []interface{} {
	ret := make([]interface{}, len(datas))
	length := -1
	for _, data := range datas {
		length ++
		if key != "" {
			ret[length] = data[key]
			continue
		}
		for _, v := range data {
			ret[length] = v
			break
		}
	}
	return ret[:length+1]
}

func listUnique(data []interface{}) []interface{} {
	newData := make([]interface{}, len(data))
	length := 0
	for _, d := range data {
		isExist := false
		for i:=0; i<length; i++ {
			if dString, ok := newData[i].(string); ok && dString == d.(string) {
				isExist = true
				break
			}
			if dInt, ok := newData[i].(int); ok && dInt == d.(int) {
				isExist = true
				break
			}
		}
		if ( ! isExist) {
			newData[length] = d
			length++
		}
	}
	newData = newData[:length]
	return newData
}

var reg = []*regexp.Regexp {
	regexp.MustCompile(`.+(?:ICP|icp)\s*备\d+号`),
	regexp.MustCompile(`.+(?:ICP|icp)\s*备.{0,2}\d+`),
	regexp.MustCompile(`.+(?:ICP|icp)\s*证`),
	regexp.MustCompile(`.+\w\d--\d+`),
	regexp.MustCompile(`.+\w\d-\d+`),
	regexp.MustCompile(`^\d+$`),
}

func IsValidIcp(icp interface{}) bool {
	i, ok := icp.(string)
	if ! ok { return false }
	if idx := strings.Index(i, ":"); idx > 0 { i = i[idx+1:] }
	i = strings.Trim(i, " ()[]")
	for _, r := range reg {
		if r.MatchString(i) { return true }
	}
	return false
}

var cacheTime time.Time
var cacheList []map[string] interface{}

func getList() ([]map[string] interface{}, error) {
	if cacheList != nil && time.Now().Sub(cacheTime) < 30 * time.Second {
		return cacheList, nil
	}
	var list []map[string] interface{}
	err := callPostAdmin(&list, "/proxy/api?biz/admin/cdomain/list", nil)
	if err == nil {
		cacheList = list
		cacheTime = time.Now()
	}
	return list, err
}

func List(c *context.Context) {
	fields := c.StringSplit("field", "-", nil)
	buckets := c.Strings("bucket", nil)
	showList := c.Bool("showList", false)
	uniqueList := c.Bool("uniqueList", true)
	uniqueField := c.String("uniqueField", "")
	limit := c.Int("limit", -1)
	statuses := c.Ints("status", nil)
	onlyIcp := c.Bool("onlyIcp", false)
	onlyNoIcp := c.Bool("onlyNoIcp", false)
	domains := c.Strings("domain", nil)
	cdns := c.Ints("cdn", nil)
	before := c.Int64("before", 864000*3)
	useCache := c.Bool("useCache", true)
	_ = useCache
	br := c.Bool("br", false)
	if showList && len(fields) != 1 {
		c.ReplyErrorInfo("unable show list and multiple field")
		return
	}
	if onlyIcp && onlyNoIcp {
		c.ReplyErrorInfo("unable to show with onlyIcp and onlyNoIcp")
		return
	}
	if len(uniqueField) > 0 && len(fields) > 0 && ! c.IsInStringArray(uniqueField, fields) {
		c.ReplyErrorInfo("field not contains uniqueField")
		return
	}
	list, err := getList()
	if err != nil {
		c.ReplyError(err)
		return
	}

	newList := make([]map[string] interface{}, len(list))
	length := 0
	for _, item := range list {
		if len(statuses) > 0 {
			pass := false
			for _, status := range statuses {
				if int(item["status"].(float64)) == status {
					pass = true
					break
				}
			}
			if ! pass { continue }
		}
		newList[length] = make(map[string] interface{}, len(item))
		isIcp := IsValidIcp(item["icp"])
		if onlyIcp && ! isIcp { continue }
		if onlyNoIcp && isIcp { continue }
		// add host
		item["host"], err = domain.GetHost(item["domain"].(string))
		if err != nil { item["host"] = item["domain"].(string) }
		if len(domains) > 0 {
			pass := false
			for _, domain := range domains {
				d := item["domain"].(string)
				if d == domain || strings.HasSuffix(d, "." + domain) {
					pass = true
					break
				}
			}
			if ! pass { continue }
		}
		if len(buckets) > 0 {
			pass := false
			for _, bucket := range buckets {
				if bucket == item["bucket"] {
					pass = true
					break
				}
			}
			if ! pass { continue }
		}
		if len(cdns) > 0 {
			pass := false
			for _, cdn := range cdns {
				if int(item["cdn"].(float64)) == cdn {
					pass = true
					break
				}
			}
			if ! pass { continue }
		}
		pass := false
		for i:=0; i<length; i++ {
			if uniqueField != "" {
				if newList[i][uniqueField].(string) == item[uniqueField].(string) {
					pass = true
					break
				}
			}
		}
		if before != 0 {
			updateTime := int64(item["update_time"].(float64))
			timeNow := time.Now().Unix()
			if before > 0 && timeNow-before > updateTime { continue }
			if before < 0 && timeNow+before < updateTime { continue }
		}
		if pass { continue }
		for k, v := range item {
			if len(fields) > 0 && ! c.IsInStringArray(k, fields) { continue }
			if v2, ok := v.(float64); ok { v = int(v2) }
			newList[length][k] = v
		}
		length ++
	}
	newList = newList[:length]
	if ! showList {
		if limit > 0 && limit < len(newList) { newList = newList[:limit] }
		if br { c.ReplyJSON(newList) }
		c.ReplyObj(newList)
	}

	// showList
	dataList := mapToList(newList, "")
	if uniqueList { dataList = listUnique(dataList) }
	if limit > 0 && limit < len(newList) { dataList = dataList[:limit] }
	if br {
		data := ""
		for _, item := range dataList {
			data += item.(string) + "<br>"
		}
		c.ReplyHtml(data)
	}
	c.ReplyObj(dataList)
}
