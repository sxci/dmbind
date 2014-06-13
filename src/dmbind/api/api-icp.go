package api

import (
	"dmbind/lib/context"
	"net/http"
	"bufio"
)

func getIcp(host string) (icp string, err error) {
	resp, err := http.Get("http://icp.aizhan.com/geticp/?host="+host+"&style=1")
	if err != nil { return }
	defer resp.Body.Close()
	buf := bufio.NewReader(resp.Body)
	ret, _, err := buf.ReadLine()
	if err != nil { return }
	icp = string(ret[16:len(ret)-3])
	if icp == "未查到备案信息" {
		icp = ""
	}
	return
}

func Icp(c *context.Context) {
	hosts := c.MustStrings("host")
	icp := make(map[string] string, len(hosts))
	for _, h := range hosts {
		i, err := getIcp(h)
		if err != nil { continue }
		icp[h] = i
	}
	c.ReplyObj(icp)
}
