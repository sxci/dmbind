package pub

import (
	"encoding/base64"
)

var RS_HOST = "http://rs.qbox.me"
var mac *Mac

func Setup(m *Mac) {
	mac = m
}

func Publish(domain, bucketName string) (err error) {
	err = New(mac).Conn.Call(nil, nil, RS_HOST + "/publish/" + encode(domain) + "/from/" + bucketName)
	return
}

func Unpublish(domain string) (err error) {
	err = New(mac).Conn.Call(nil, nil, RS_HOST + "/unpublish/" + encode(domain))
	return
}

func encode(u string) string {
	return base64.URLEncoding.EncodeToString([]byte(u))
}

