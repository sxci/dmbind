package cnc

import (
	"fmt"
	"math/rand"
	"time"
)

type RelativeDomain struct {
	Id     string `json:"id"`
	Domain string `json:"domain"`
	// IsUse      bool
	// UseVersion string
	// Domains    []string
	// UseTime    time.Time
}

var relativeDomain = []*RelativeDomain{
	{"96795", ".aisheji.org"},
	{"96779", ".dian-ying.org"},
	{"98307", "ws-header-antileech.qiniudn.com"},
}

var r *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))

func SetRel(rel []*RelativeDomain) {
	relativeDomain = rel
	fmt.Println(getIdleRelativeId())
}

func getIdleRelativeId() (id string, ok bool) {
	l := len(relativeDomain)
	fmt.Println(l)
	id = relativeDomain[r.Intn(l)].Id
	ok = true
	return
}

func getDomainsFromVersion(version string) (domains []string) {
	domains = []string{}
	return
}

func finishRelativeId(id string) {
}

func fillVersion(id string, version string, domains []string) {
}
