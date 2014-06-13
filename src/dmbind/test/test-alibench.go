package test

import (
	"sync"
	"time"
	"net/url"
)

import (
	"dmbind/lib/alibench"
	"dmbind/lib/pub"
	"dmbind/lib/context"
)

var (
	globalQueue = NewQueue()
	testPrefix = "dmbind.test."
)

func init() {
	globalQueue.Run(31)
}

func Alibench(c *context.Context) {
	host := c.MustString("domain")
	testUrl := c.MustString("testUrl")
	bucket := c.MustString("bucket")
	testHost := testPrefix + host

	result, err := MustTest(testHost, testUrl, bucket)
	c.ReplyIfError(err)
	c.ReplyObj(result.Success)
}

type TestResult struct {
	Info     string `json:"info"`
	Success  bool   `json:"success"`
	TestPath string `json:"test_path"`
}

func MustTest(testDomain, testUrl, bucket string) (result TestResult, err error) {
retry:
	result, err = func() (ret TestResult, err error) {
		err = pub.Publish(testDomain, bucket)
		if err != nil { return }
		defer pub.Unpublish(testDomain)

		err = context.Call(&ret, domain.AliTest, url.Values {
			"url": {testUrl},
			"host": {testDomain},
		})
		return
	}()
	
	if err != nil {
		if err.Error() != "too busy" { return }
		globalQueue.AddItem().WaitForNotify()
		goto retry
	}
	return
}

// ----------------------------------------------------------------------------

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
