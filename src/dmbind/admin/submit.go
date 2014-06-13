package admin

import (
	"strconv"
	"strings"
	"net/url"
)

import (
	// "dmbind/cnc"
	"dmbind/mail"
	"dmbind/domain"
	"dmbind/lib/context"
	// "dmbind/dnspod"
	// "dmbind/tx"
)

var icpNotifyMail []string
var localhost string

func SetupIcpNotifyMail(mail []string) {
	icpNotifyMail = mail
}

func SetupLocalhost(host string) {
	localhost = host
}

func Filter(c *context.Context) {
	var successMap []DomainInfo
	c.ReplyIfError(context.Call(&successMap, List, url.Values {
		"status": {"2", "4"},
		"field": {"host-cdn"},
		"before": {"0"},
	}))
	successList := make(map[string] struct{}, len(successMap))
	for _, succ := range successMap {
		successList[succ.Host + strconv.Itoa(succ.Cdn)] = struct{}{}
	}
	var waitingMap []DomainInfo
	c.ReplyIfError(context.Call(&waitingMap, List, url.Values {
		"status": {"3"},
		"field": {"host-cdn"},
	}))
	waitingList := make(map[string] struct{}, len(waitingMap))
	for _, wait := range waitingMap {
		waitingList[wait.Host + strconv.Itoa(wait.Cdn)] = struct{}{}
	}
	var list []DomainInfo
	c.ReplyIfError(context.Call(&list, List, url.Values {
		"status": {"1"},
		"before": {"2592000"},
		"field": {"host-cdn"},
		"uniqueField": {"host"},
		"onlyIcp": {"1"},
	}))
	var toDeny []string
	c.ReplyIfError(context.Call(&toDeny, List, url.Values {
		"status": {"1"},
		"before": {"2592000"},
		"field" : {"host"},
		"uniqueList": {"1"},
		"showList": {"1"},
		"onlyNoIcp": {"1"},
	}))

	toSuccess, lengthSuccess := make([]string, len(list)), 0
	toWait, lengthWait := make([]string, len(list)), 0
	unknown, lengthUnknown := make([]string, len(list)), 0
	for _, item := range list {
		if _, ok := successList[item.Host + strconv.Itoa(item.Cdn)]; ok {
			toSuccess[lengthSuccess] = item.Host
			lengthSuccess++
			continue
		}
		if _, ok := waitingList[item.Host + strconv.Itoa(item.Cdn)]; ok {
			toWait[lengthWait] = item.Host
			lengthWait++
			continue
		}
		if InArray(item.Host, toDeny) { continue }
		unknown[lengthUnknown] = item.Host
		lengthUnknown++
	}
	toSuccess = toSuccess[:lengthSuccess]
	toWait = toWait[:lengthWait]
	unknown = unknown[:lengthUnknown]
	var domainIcp map[string] []string
	c.ReplyIfError(context.Call(&domainIcp, domain.Icp, url.Values {
		"select": {"1"},
		"domain": append(unknown, toDeny...),
	}))
	newDeny := []string {}
	for _, un := range domainIcp["unknown"] {
		if InArray(un, toDeny) {
			newDeny = append(newDeny, un)
		}
	}
	ret := map[string] []string {
		"success": toSuccess,
		"deny": newDeny,
		"wait": toWait,
		"unknown": domainIcp["unknown"],
		"icped": domainIcp["icp"],
	}
	c.ReplyObj(ret)
}

func InArray(s string, lib []string) bool {
	for _, i := range lib { if i == s { return true } }
	return false
}

func FilterSubmit(c *context.Context) {
	actions := c.StringSplit("action", "-", []string{"success","deny","wait","icped","unknown"})
	cdns := c.StringSplit("cdn", "-", []string{"0", "1"})
	icpBufferCount := c.Int("buffer", 1)
	var ret map[string] []string
	c.ReplyIfError(context.Call(&ret, Filter, nil))
	for key, _ := range ret {
		if InArray(key, actions) { continue }
		delete(ret, key)
	}
	if len(ret["icped"]) >= icpBufferCount {
		// cnc
		if InArray("0", cdns) {
			var cncDomains []string
			err := context.Call(&cncDomains, List, url.Values {
				"field": {"host"},
				"domain": ret["icped"],
				"showList": {"1"},
				"uniqueList": {"1"},
				"cdn": {"0"},
				"status": {"1"},
			})
			if err != nil {
				c.ErrorWithTag("icped.cnc.filter", err)
			} else if len(cncDomains) > icpBufferCount {
				/*
				err = context.Call(nil, cnc.Submit, url.Values {
					"domain": cncDomains,
				})
				*/
				if err != nil {
					c.ErrorWithTag("icped.cnc.submit", err)
				} else {
					c.Info("submit domains to cnc success, switch to wait,", cncDomains)
					ret["wait"] = append(ret["wait"], cncDomains...)
				}
			} else {
				c.Info("icped.cnc.submit", "got", cncDomains, "waiting", icpBufferCount)
			}
		}

		// txc
		if InArray("1", cdns) {
			var txcDomains []string
			err := context.Call(&txcDomains, List, url.Values {
				"field": {"host"},
				"domain": ret["icped"],
				"showList": {"1"},
				"uniqueList": {"1"},
				"cdn": {"1"},
				"status": {"1"},
			})
			if err != nil {
				c.ErrorWithTag("icped.txc.filter", err)
			} else if len(txcDomains) > icpBufferCount {
				// tmp submit to ws
				// err = context.Call(nil, tx.Submit, url.Values {
				/*
				err = context.Call(nil, cnc.Submit, url.Values {
					"domain": txcDomains,
				})
				*/
				if err != nil {
					c.ErrorWithTag("icped.txc.submit", err)
				} else {
					c.Info("submit domains to txc success, switch to wait,", txcDomains)
					ret["wait"] = append(ret["wait"], txcDomains...)
				}
			}
		}
	}

	if err := onSuccess(c, ret["success"]); err != nil {
		c.ErrorWithTag("pass", err, ret["success"])
	}

	if len(ret["wait"]) > 0 {
		err := context.Call(nil, Status, url.Values {
			"status": {"3"},
			"domain": ret["wait"],
			"from": {"1"},
		})
		if err != nil {
			c.ErrorWithTag("wait", err, ret["wait"])
		} else {
			c.Info("set domains to wait", ret["wait"])
		}
	}

	if len(ret["deny"]) > 0 {
		err := context.Call(nil, Deny, url.Values {
			"domain": ret["deny"],
			"from": {"1"},
		})
		if err != nil {
			c.ErrorWithTag("deny", err)
		} else {
			c.Info("set domains to deny", ret["deny"])
		}
	}

	if len(ret["unknown"]) > 0 {
		path := "http://"+localhost+"/admin/icp?domain=" + strings.Join(ret["unknown"], "&domain=")
		err := context.Call(nil, mail.Send, url.Values {
			"profile": {"noreply"},
			"to": icpNotifyMail,
			"subject": {"未知备案信息域名"},
			"content": {strings.Join(ret["unknown"], "\n") + "\n" + path},
		})
		if err != nil {
			c.ErrorWithTag("unknown", err)
		} else {
			c.Info("sending mail to notify unknow domain", ret["unknown"])
		}
	}
	c.ReplyObj(true)

}

// tmp
var txStartTime int64 = 1381207640
func getDnspodBuckets(domains []string) (buckets []string, err error) {
	var ret []DomainInfo
	err = context.Call(&ret, List, url.Values {
		"domain": domains,
		"status": {"1"},
		"field": {"create_time-bucket"},
		"cdn": {"1"},
	})
	if err != nil { return }
	buckets = make([]string, len(ret))
	length := 0
	for _, item := range ret {
		if item.CreateTime < txStartTime { continue }
		buckets[length] = item.Bucket
		length++
	}
	buckets = buckets[:length]
	return
}

func onSuccess(c *context.Context, domains []string) (err error) {
	if len(domains) <= 0 { return }

	// tmp
	// buckets, err := getDnspodBuckets(domains)
	// if err == nil {
		// err = context.Call(nil, dnspod.DnspodCName, url.Values {
			// "bucket": buckets,
			// "raiseIfOne": {"1"},
		// })
		// if err != nil { c.Error(err) }
	// } else {
		// c.Error(err)
	// }

	err = context.Call(nil, Pass, url.Values {
		"domain": domains,
		"from": {"1"},
	})
	if err != nil { return }
	c.Info("set domains to pass", domains)
	return
}
