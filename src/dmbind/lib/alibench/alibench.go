package domain

import (
	"fmt"
	"time"
	"bytes"
	"errors"
	"net/url"
	"strings"
	"encoding/json"
)

import (
	"dmbind/lib/context"
	"dmbind/lib/fetcher"
)

var TestTime = time.Duration(60)

type TestResult struct {
	Info     string `json:"info"`
	Success  bool   `json:"success"`
	TestPath string `json:"test_path"`
}

func AliTest(c *context.Context) {
	testUrl := c.MustString("url")
	host := c.MustString("host")
	c.Info("creating task: " + host + ", url:", testUrl, "...")
	f, id, isBusy, err := newTask(testUrl, host)
	c.ReplyIfError(err)
	if isBusy {
		c.ReplyErrorInfo("too busy")
		return
	}
	c.Info("task created! host:", host, "url:", testUrl, "id:", id)
	time.Sleep(TestTime*time.Second)
	c.Info("stoping task... host:", host, "urls:", testUrl, "id:", id)
	var ret map[string] interface{}
	c.ReplyIfError(f.CallPostForm(&ret, "/ajax-stop.php", url.Values {
		"task_ids": {id},
	}))
	if ret["code"].(float64) != 0 {
		c.ReplyErrorInfo("unexcept error(stop ajax): " + id)
		return
	}
fetchData:
	_, body, err := f.Get("/rp/" + id)
	if err != nil { return }

	if idx := bytes.Index(body, []byte("TASK_DATA")); idx > 0 {
		body = body[idx+12:]
	} else {
		err = errors.New("unexcept result")
		return
	}

	idx := bytes.Index(body, []byte("];\n_gaq.push"))
	if idx <= 0 {
		c.Info("get unexcept result, refetch: " + host + ", " + id)
		time.Sleep(time.Second)
		goto fetchData
	}

	body = body[:idx+1]
	var rets []BenchRet
	c.ReplyIfError(json.Unmarshal(body, &rets))
	successCount := 0
	for _, ret := range rets { if ret.Status == "200" { successCount++ } }
	c.ReplyObj(TestResult {
		Info: fmt.Sprintf(`{"id": "%s", "host": "%s", "success": "%v", "total": "%d"}`, id, host, successCount, len(rets)),
		Success: successCount*100/len(rets) > 40 && len(rets)>100,
		TestPath: alibenchUrl(id),
	})
}

type BenchRet struct {
	ISP string      `json:"isp"`
	City string     `json:"node_city"`
	Error string    `json:"error"`
	Status string   `json:"http_response_code"`
	Province string `json:"node_province"`
}

type TaskRet struct {
	Code int    `json:"code"`
	Data string `json:"data"`
	Env  string `json:"env"`
}

func newTask(furl, host string) (f *fetcher.Fetcher, id string, isBusy bool, err error) {
	f = fetcher.NewFetcher("alibench.com")
	f.Get("/")
	data := url.Values {
		"task_from": {"self"},
		"is_pk": {"false"},
		"target": {furl},
		"ac": {"http"},
		"http_host": {host},
	}

	var ret TaskRet
	err = f.CallPostForm(&ret, "/new_task.php", data)
	if err != nil { return }
	if ret.Code != 0 {
		if strings.Contains(ret.Data, "太高") {
			isBusy = true
			return
		}
		err = errors.New(ret.Data)
		return
	}

	id = ret.Data[4:]
	return
}

func alibenchUrl(id string) string {
	return "http://alibench.com/rp/" + id
}
