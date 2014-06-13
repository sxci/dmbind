package context

import (
	"fmt"
	"log"
	"time"
	"errors"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"net/url"
	"net/http"
	"net/http/httptest"
	"encoding/json"
)

type Context struct {
	w http.ResponseWriter
	Req *http.Request
	Tag string
}

func NewFakeContext(method string, path string, params url.Values) (c *Context) {
	req, err := http.NewRequest(method, path, strings.NewReader(params.Encode()))
	if err != nil { panic(err) }
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	c = &Context {httptest.NewRecorder(), req, path}
	return
}

func NewContext(w http.ResponseWriter, req *http.Request) (c *Context) {
	c = &Context {
		w: w,
		Req: req,
	}
	return
}

func Wrap(f func(*Context)) func(w http.ResponseWriter, req *http.Request) {
	return func (w http.ResponseWriter, req *http.Request) {
		now := time.Now().UnixNano()
		ctx := NewContext(w, req)
		ctx.Tag = GetFunctionName(f)
		defer func() {
			obj := recover()
			defer ctx.Info("request time", (time.Now().UnixNano() - now) / 10e6, "ms")
			if obj == nil { return }
			switch obj.(type) {
			case string:
				if obj == "" { return }
				panic(obj)
			default:
				panic(obj)
			}
		}()
		f(ctx)
	}
}

func (c *Context) Write(data []byte) (int, error) {
	return c.w.Write(data)
}

func (c *Context) Reply(data []byte) {
	c.w.Header().Set("Content-Type", "text/html; charset=UTF-8")
	c.Write(data)
	panic("")
}

func (c *Context) ReplyText(text string) {
	c.w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
	c.Write([]byte(text))
	panic("")
}

func (c *Context) ReplyHtml(html string) {
	c.w.Header().Set("Content-Type", "text/html; charset=UTF-8")
	c.Write([]byte(html))
	panic("")
}

func (c *Context) ReplyErrorInfo(info string) {
	c.Error(info)
	c.w.WriteHeader(400)
	c.w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
	c.Write([]byte(info))
	panic("")
}

func (c *Context) ReplyError(err error) {
	c.Error(err)
	c.w.WriteHeader(400)
	c.w.Header().Set("Content-type", "text/plain; charset=UTF-8")
	c.Write([]byte(err.Error()))
	panic("")
}

func (c *Context) ReplyObj(ret interface{}) {
	c.w.WriteHeader(200)
	c.w.Header().Set("Content-Type", "application/json")
	encode := json.NewEncoder(c.w)
	err := encode.Encode(ret)
	c.ReplyIfError(err)
	panic("")
}

func (c *Context) ReplyJSON(ret interface{}) {
	c.w.WriteHeader(200)
	c.w.Header().Set("Content-Type", "application/json")
	b, err := json.MarshalIndent(&ret, "", "\t")
	c.ReplyIfError(err)
	c.w.Write(b)
	panic("")
}

func (c *Context) FakeString() string {
	recoreder, _ := c.w.(*httptest.ResponseRecorder)
	return recoreder.Body.String()
}

func (c *Context) GetResponseRecoreder() *httptest.ResponseRecorder {
	return c.w.(*httptest.ResponseRecorder)
}

func (c *Context) Println(o interface{}) {
	fmt.Printf("%+v\n", o)
}

func (c *Context) String(field string, def string) string {
	data := c.Strings(field, []string {def})
	if len(data) == 0 { return def }
	return data[0]
}

func (c *Context) Strings(field string, def []string) []string {
	urlQuery := c.Req.URL.Query()
	data, ok := urlQuery[field]
	if ! ok {
		if c.Req.Method != "POST" { return def }
		err := c.Req.ParseForm()
		if err != nil { return def }
		ret, ok := c.Req.Form[field]
		if ! ok { return def }
		return ret
	}
	return data
}

func (c *Context) Bool(field string, def bool) bool {
	defString := "0"
	if def { defString = "1" }
	ret := c.String(field, defString)
	return ret == "1"
}

func (c *Context) Int(field string, def int) int {
	defString := strconv.Itoa(def)
	data := c.String(field, defString)
	ret, err := strconv.Atoi(data)
	if err != nil { return def }
	return ret
}

func (c *Context) Ints(field string, def []int) []int {
	retStr := c.Strings(field, nil)
	if retStr == nil { return def }
	ret := make([]int, len(retStr))
	length := 0
	for _, i := range retStr {
		if i == "" { continue }
		ret[length], _ = strconv.Atoi(i)
		length++
	}
	return ret[:length]
}

func (c *Context) Int64(field string, def int64) int64 {
	defString := strconv.FormatInt(def, 10)
	data := c.String(field, defString)
	ret, err := strconv.ParseInt(data, 10, 0)
	if err != nil { return def }
	return ret
}

func (c *Context) StringSplit(field string, spch string, def []string) []string {
	ret := c.String(field, "")
	if ret == "" { return def }
	return strings.Split(ret, spch)
}

func (c *Context) IsInStringArray(s string, b []string) bool {
	for _, bs := range b { if s == bs { return true } }
	return false
}

func (c *Context) FakeGetHeader(field string) string {
	header := c.GetResponseRecoreder().Header()
	return header.Get(field)
}

func GetFunctionName(i interface{}) string {
    return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}

func Get(f func(*Context), params url.Values) (ret string, err error) {
	ctx := NewFakeContext("GET", "/?" + params.Encode(), nil)
	ctx.Tag = GetFunctionName(f)
	defer func() {
		recover()
		if ctx.GetResponseRecoreder().Code / 100 != 2 {
			err = errors.New(ctx.FakeString())
			return
		}
		ret = ctx.FakeString()
	}()
	f(ctx)
	return

}

func Call(ret interface{}, f func(*Context), params url.Values) (err error) {
	ctx := NewFakeContext("POST", "/", params)
	ctx.Tag = GetFunctionName(f)
	defer func() {
		recover()
		if ctx.GetResponseRecoreder().Code / 100 != 2 {
			err = errors.New(ctx.FakeString())
			return
		}
		if ret == nil { return }
		switch ctx.FakeGetHeader("Content-Type") {
		case "application/json":
			err = json.Unmarshal([]byte(ctx.FakeString()), ret)
			if err != nil {
				err = errors.New(err.Error() + ", detail:" + ctx.FakeString())
				return
			}
		default:
			data := ctx.FakeString()
			if len(data) > 2 && data[0] == '{' && data[len(data)-1] == '}' {
				err = json.Unmarshal([]byte(data), ret)
				if err != nil {
					err = errors.New(err.Error() + ", detail:" + ctx.FakeString())
					return
				}
				return
			}
		}
	}()
	f(ctx)
	return
}

func (c *Context) MustString(field string) string {
	s := c.String(field, "")
	if s != "" { return s }
	c.ReplyErrorInfo("miss " + field)
	panic("")
}

func (c *Context) MustStrings(field string) []string {
	s := c.Strings(field, nil)
	if len(s) > 0 { return s }
	c.ReplyErrorInfo("miss " + field)
	panic("")
}

func (c *Context) IsPost() bool {
	return c.Req.Method == "POST"
}

func (c *Context) ReplyIfError(err error) {
	if err == nil { return }
	c.ReplyError(err)
	panic("")
}

func (c *Context) Redirect(path string) {
	c.w.Header().Set("Location", path)
	c.w.WriteHeader(302)
	panic("")
}

func (c *Context) Info(o ...interface{}) {
	o = append([]interface{}{"[INFO][" + c.Tag + "]"}, o...)
	log.Println(o...)
}

func (c *Context) Error(o ...interface{}) {
	o = append([]interface{}{"[ERROR][" + c.Tag + "]"}, o...)
	log.Println(o...)
}

func (c *Context) ErrorWithTag(tag string, o...interface{}) {
	o = append([]interface{}{tag, "->"}, o...)
	log.Println(o...)
}
