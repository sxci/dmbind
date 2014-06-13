package admin

import (
	"sync"
	"time"
	"strings"
	"net/url"
	"strconv"
	"math/rand"
)

import (
	"dmbind/lib/pub"
	"dmbind/domain"
	"dmbind/lib/context"
)

var globalQueue = NewQueue()
var Hosts map[int][]string
var TestFailNotifyMail = []string {"chenye@qiniu.com"}
var bindBucket = []string{"demo", "txc-test"}
var qiniuSourceUrl = []string {"http://ws1.source.qbox.me", "http://tx1.source.qbox.me"}

func SetupTest(failNotifyMail, buckets, cdnSourceUrl []string) {
	TestFailNotifyMail = failNotifyMail
	bindBucket = buckets
	qiniuSourceUrl = cdnSourceUrl
}

func init() {
	maxInt := 128
	cncId := 0
	txcId := 1
	Hosts = make(map[int] []string, 2)
	for i:=0; i<2; i++ {
		Hosts[i] = make([]string, maxInt)
	}
	for i:=0; i<maxInt; i++ {
		Hosts[cncId][i] = "http://ali-test" + strconv.Itoa(i+1) + ".qiniudn.com/a.py"
		Hosts[txcId][i] = "http://ali-txc-" + strconv.Itoa(i+1) + ".u.qiniudn.com/a.py"
	}
	rand.Seed(time.Now().Unix())
	globalQueue.Run(31)
}

func AutoTest(c *context.Context) {
	var domains []string
	c.ReplyIfError(context.Call(&domains, List, url.Values {
		"status": {"3"},
		"showList": {"1"},
		"field": {"host"},
		"uniqueList": {"1"},
	}))
	var tmpdm string
	for i:=0; i<len(domains)/2; i++ {
		j := len(domains)-1-i
		if i==j { break }
		tmpdm = domains[i]
		domains[i] = domains[j]
		domains[j] = tmpdm
	}
	var ret interface{}
	c.ReplyIfError(context.Call(&ret, BatchTestDomains, url.Values {
		"domain": domains,
	}))
	c.ReplyObj(ret)
}

func BatchTestDomains(c *context.Context) {
	domains := c.Strings("domain", nil)
	if len(domains) >  0 || c.IsPost() {
		domainStr := ""
		if len(domains) != 0 {
			domainStr = strings.Join(domains, "\n")
		} else {
			domainStr = c.MustString("domain_str")
		}
		domains := strings.Split(strings.TrimSpace(domainStr), "\n")
		for idx, domain := range domains {
			if strings.HasPrefix(domain, "*.") { domain = domain[2:] }
			domains[idx] = strings.TrimSpace(domain)
		}
		var ret interface{}
		c.ReplyIfError(context.Call(&ret, TestDomains, url.Values {
			"domain": domains,
		}))
		c.ReplyObj(ret)
		return
	}
	html := `
	<form action="" method="POST">
	<textarea type="text" value="" name="domain_str" ></textarea>
	<input type="submit" />
	</form>`
	c.ReplyHtml(html)
}

func TestDomains(c *context.Context) {
	hosts := c.MustStrings("domain")
	rets := make([]interface{}, len(hosts))
	ret_chan := make(chan interface{}, len(hosts))
	for _, host := range hosts {
		go func(host string) {
			var ret interface{}
			err := context.Call(&ret, TestDomain, url.Values {
				"domain": {host},
			})
			if err != nil {
				ret_chan <- err.Error()
			} else {
				ret_chan <- ret
			}
		}(host)
	}
	for idx, _ := range hosts {
		rets[idx] = <-ret_chan
	}
	c.ReplyObj(rets)
}

// 临时的方案，用于将同兴的域名使用网宿的方法去测试
func TestTxToWsDomain(c *context.Context) {
	host := c.MustString("domain")
	var domains []string
	c.ReplyIfError(context.Call(&domains, List, url.Values{
		"domain": {host},
		"field": {"domain"},
		"showList": {"1"},
		"uniqueList": {"1"},
		"status": {"3"},
		"cdn": {"1"}, //tx
	}))
	if len(domains) == 0 {
		c.ReplyErrorInfo("invalid domain that not tx domain")
		return
	}

	testUrl := randHost(0)
	testDomain := "test-12jvb03nf1lf." + host
	result, err := MustTest(testDomain, testUrl, bindBucket[0])
	c.ReplyIfError(err)
	if result.Success {
		// err := context.Call(nil, dnspod.CdnToTx, url.Values {
			// "domain": domains,
		// })
		// if err != nil { c.Error(err) }
		c.ReplyIfError(context.Call(nil, Pass, url.Values{
			"domain": {host},
			"cdn": {"1"},
		}))
	} else {
		c.Error("test domain", host, "fail! ", result)
	}
	c.ReplyObj(result)
}

// check is in waiting list
// bind domain
// test domain
// unbind domain
//
func TestDomain(c *context.Context) {
	host := c.MustString("domain")
	must := c.Bool("must", false)
	tx := c.Bool("tx", false)
	var ret []DomainInfo
	c.ReplyIfError(context.Call(&ret, List, url.Values {
		"domain": {host},
		"status": {"3"},
	}))
	cdnId := 0
	if len(ret) == 0 {
		if must {
			if tx { cdnId = 1 }
		} else {
			c.ReplyErrorInfo("invalid domain:" + host)
			return
		}
	} else {
		cdnId = ret[0].Cdn
	}

	// tmp
	if cdnId == 1 && ! tx {
		var ret interface{}
		c.ReplyIfError(context.Call(&ret, TestTxToWsDomain, url.Values {"domain": {host}}))
		c.ReplyObj(ret)
		return
	}

	testUrl := randHost(cdnId)
	if ! must {
		var ret []DomainInfo
		c.ReplyIfError(context.Call(&ret, List, url.Values {
			"domain": {host},
			"status": {"3"},
		}))
		if len(ret) == 0 {
			c.ReplyErrorInfo("domain " + host + " not in waiting list")
			return
		}
	}
	testDomain := "test-12jvb03nf1lf." + host

	result, err := MustTest(testDomain, testUrl, bindBucket[cdnId])
	c.ReplyIfError(err)
	if result.Success {
		c.ReplyIfError(context.Call(nil, Pass, url.Values {
			"domain": {host},
			"cdn": {strconv.Itoa(cdnId)},
		}))
	} else {
		c.Error("test domain", host, "fail,", result)
	}
	c.ReplyObj(result)
}

func MustTest(testDomain, testUrl, bucket string) (result domain.TestResult, err error) {
retry:
	// bind domain
	err = pub.Publish(testDomain, bucket)
	if err != nil { return }
	err = context.Call(&result, domain.AliTest, url.Values {
		"url": {testUrl},
		"host": {testDomain},
	})
	// unbind domain
	pub.Unpublish(testDomain)
	if err != nil {
		if err.Error() != "too busy" { return }
		globalQueue.AddItem().WaitForNotify()
		goto retry
	}
	return
}

func randHost(cdnId int) string { return Hosts[cdnId][rand.Int() & (len(Hosts[0])-1)] }

// ----------------------------------------------------------------------

type Queue chan *QueueItem

func NewQueue() Queue {
	return Queue(make(chan *QueueItem, 1024))
}

func (q Queue) Run(duration int) {
	go func() {
		for {
			i := <- q
			i.wakeUp()
			time.Sleep(time.Duration(duration) * time.Second)
		}
	}()
}

func (q Queue) AddItem() (qi *QueueItem){
	var wg sync.WaitGroup
	qi = &QueueItem{len(q), wg}
	qi.wg.Add(1)
	q <- qi
	return
}

type QueueItem struct {
	id int
	wg sync.WaitGroup
}

func (qi *QueueItem) wakeUp() {
	qi.wg.Done()
}

func (qi *QueueItem) WaitForNotify() {
	qi.wg.Wait()
}
