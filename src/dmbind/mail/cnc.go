package mail

import (
	"os"
	"errors"
	"strings"
	"io/ioutil"
	"encoding/json"
	"dmbind/lib/context"
)

var (
	AlreadySwitchFile = "mailTmp2.txt"
	mailServerProfile MailProfile
)

func SetupMailServerProfile(profileName string) (err error) {
	profile, ok := profiles[profileName]
	if ! ok {
		err = errors.New("miss profile " + profileName)
		return
	}
	mailServerProfile = profile
	return
}

func GetCncTestUrl(c *context.Context) {
	user, pswd := mailServerProfile.User, mailServerProfile.Pswd
	mailHost := mailServerProfile.Server + ":" + mailServerProfile.Port
	s, err := NewMailServer(user, pswd, mailHost)
	c.ReplyIfError(err)
	infos, err := s.GetCncTestInfoBatch(c)
	c.ReplyIfError(err)
	c.ReplyObj(infos)
}

func SendMailNotify(all []string, url string) {
	println("sending notify")
}

func GetCncTestInfo(c *context.Context, ret Mail) (domains []string, url string, err error) {
	body := string(ret.All[len(ret.All)-1])
	body = strings.Replace(body, "<br />", "", -1)
	body = strings.Replace(body, "&nbsp;", " ", -1)
	body = strings.Replace(body, "&amp;", "&", -1)
	
	idx := strings.Index(body, "域名:")
	if idx < 0 {
		err = errors.New("error in match domains")
		return
	}
	data := body[idx:]
	idx = strings.Index(data, "\n")
	if idx < 0 {
		err = errors.New("error in match domains, miss \\n")
		return
	}
	data = data[idx+1:]
	data = data[:strings.Index(data, "\n")]
	
	data = strings.Trim(data, string([]byte{160}) + "\r\n ")
	allDomains := strings.Split(data, ";")
	domains = make([]string, len(allDomains) / 2)
	if len(allDomains) % 2 == 0 {
		for i:=0; i<len(allDomains); i+=2 {
			domains[i/2] = strings.Replace(allDomains[i], "..", ".", -1)
		}
	}
	
	idx = strings.Index(body, "http://read.chinanetcenter.com/myeasy")
	if idx <= 0 {
		err = errors.New("Can't find alias name url")
		return
	}
	aliasNameUrl := body[idx:]
	idx = strings.Index(aliasNameUrl, "\r\n")
	if idx < 0 {
		err = errors.New("Can't find alias name url(end)")
		return
	}
	aliasNameUrl = aliasNameUrl[:idx]
	if idx := strings.Index(aliasNameUrl, `"`); idx > 0 {
		aliasNameUrl = aliasNameUrl[:idx]
	}
	if idx := strings.Index(aliasNameUrl, `<`); idx > 0 {
		aliasNameUrl = aliasNameUrl[:idx]
	}
	url = aliasNameUrl
	return
}

type CncInfo struct {
	Domains []string `json:"domains"`
	Url     string   `json:"url"`
}

func (s *MailServer) GetCncTestInfoBatch(c *context.Context) (infos []CncInfo, err error) {
	rets, err := s.FindMail("网宿加速服务测试指南", s.prev)
	if err != nil { return }
	infos = make([]CncInfo, len(rets))
	length := 0
	for _, ret := range rets {
		domains, url, err1 := GetCncTestInfo(c, ret)
		if err1 != nil {
			continue
		}
		infos[length] = CncInfo {domains, url}
		length++
	}
	infos = infos[:length]
	return
}

func StoreObj(obj interface{}) (err error) {
	d, err := json.Marshal(obj)
	if err != nil { return }
	err = ioutil.WriteFile(AlreadySwitchFile, d, os.ModePerm)
	return
}

func RestoreObj(obj interface{}) (err error) {
	ret, err := ioutil.ReadFile(AlreadySwitchFile)
	if err != nil { return }
	err = json.Unmarshal(ret, obj)
	return
}
