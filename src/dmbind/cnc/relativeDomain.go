package cnc

import (
	"sync"
	"time"
)

type RelativeDomain struct {
	Id         string
	Domain     string
	IsUse      bool
	UseVersion string
	Domains    []string
	UseTime    time.Time
}

var relativeDomain = []*RelativeDomain{
	{"96795", ".aisheji.org", false, "", nil, time.Time{}},
	{"96779", ".dian-ying.org", false, "", nil, time.Time{}},
}
var mutex sync.Mutex

func getIdleRelativeId() (id string, ok bool) {
	mutex.Lock()
	defer mutex.Unlock()
	for _, i := range relativeDomain {
		if !i.IsUse {
			i.IsUse = true
			id = i.Id
			i.UseTime = time.Now()
			ok = true
			return
		}
	}

	now := time.Now()
	for _, i := range relativeDomain {
		if now.Sub(i.UseTime) > 24*time.Hour {
			id = i.Id
			i.UseTime = time.Now()
			ok = true
			return
		}
	}
	return
}

func getDomainsFromVersion(version string) (domains []string) {
	mutex.Lock()
	defer mutex.Unlock()
	for _, i := range relativeDomain {
		if i.UseVersion == version {
			return i.Domains
		}
	}
	return
}

func finishRelativeId(id string) {
	mutex.Lock()
	defer mutex.Unlock()
	for _, i := range relativeDomain {
		if i.Id == id {
			i.IsUse = false
			i.UseVersion = ""
			return
		}
	}
}

func fillVersion(id string, version string, domains []string) {
	mutex.Lock()
	defer mutex.Unlock()
	for _, i := range relativeDomain {
		if i.Id == id && i.IsUse {
			i.UseVersion = version
			i.Domains = domains
			return
		}
	}
}
