package cnc

import (
	"time"
	"sync"
)

type RelativeDomain struct {
	Id string
	Domain string
	IsUse  bool
	UseVersion string
	Domains []string
	UseTime time.Time
}

var relativeDomain = []*RelativeDomain {
	{"80470", ".noniu.com", false, "", nil, time.Time{}},
	{"75653", ".zywxrg.com", false, "", nil, time.Time{}},
	{"75813", ".banbaor.com", false, "", nil, time.Time{}},
	{"75731", ".endjia.com", false, "", nil, time.Time{}},
	{"75729", ".52zkd.com", false, "", nil, time.Time{}},
	{"75609", ".chuang1.com", false, "", nil, time.Time{}},
	{"70215", ".wzk8.com", false, "", nil, time.Time{}},
	{"72641", ".fuyouwu.com", false, "", nil, time.Time{}},
	{"75087", ".yujianin.com", false, "", nil, time.Time{}},
	{"75179", ".doubaoer.com", false, "", nil, time.Time{}},
	{"75181", ".hxtan.com", false, "", nil, time.Time{}},
	{"75536", ".5zhl.com", false, "", nil, time.Time{}},
}
var mutex sync.Mutex

func getIdleRelativeId() (id string, ok bool) {
	mutex.Lock()
	defer mutex.Unlock()
	for _, i := range relativeDomain {
		if ! i.IsUse {
			i.IsUse = true
			id = i.Id
			i.UseTime = time.Now()
			ok = true
			return
		}
	}

	now := time.Now()
	for _, i := range relativeDomain {
		if now.Sub(i.UseTime) > 24 * time.Hour {
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
