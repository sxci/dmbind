package api

import (
	"dmbind/lib/context"
	"dmbind/cnc"
	"dmbind/admin"
	"net/url"
)

func SubmitAvailable(c *context.Context) {
	var domains map[string] string
	err := context.Call(&domains, AdminAvailableDomain, nil)
	c.ReplyIfError(err)
	var unicpDomains []string
	var icpedDomains []string
	for domain, icp := range domains {
		if icp == "" {
			unicpDomains = append(unicpDomains, domain)
		} else {
			icpedDomains = append(icpedDomains, domain)
		}
	}
	var ret cnc.SafeAddRet
	err = context.Call(&ret, Add, url.Values {
		"domain": icpedDomains,
	})
	ret.Unicp = append(ret.Unicp, unicpDomains...)
	c.ReplyIfError(err)
	c.ReplyObj(ret)
}

func Add(c *context.Context) {
	domains := c.MustStrings("domain")
	src := c.String("src", "")
	var addRet cnc.SafeAddRet

	var allDomains []string
	for _, dm := range domains {
		allDomains = append(allDomains, dm, "." + dm)
	}
	err := context.Call(&addRet, cnc.AddSafe, url.Values {
		"domain": allDomains,
		"src": {src},
	})
	c.ReplyIfError(err)
	err = context.Call(nil, admin.Status, url.Values {
		"domains": domains,
		"status": {"3"},
	})
	c.ReplyIfError(err)
	c.ReplyJSON(addRet)
}

func DeployPassed(c *context.Context) {
	versions := c.MustStrings("version")
	c.ReplyObj(versions)
}

func AddOne(c *context.Context) {
	// unique domains
	domains := c.MustStrings("domain")
	src := c.String("src", "")
	var addRet cnc.SafeAddRet

	var allDomains []string
	for _, dm := range domains {
		allDomains = append(allDomains, dm, "." + dm)
	}
	err := context.Call(&addRet, cnc.AddSafe, url.Values {
		"domain": allDomains,
		"src": {src},
	})
	c.ReplyIfError(err)
	version := addRet.Ret.Version
	var ret cnc.RetSave
	err = context.Call(&ret, cnc.PassCheck, url.Values {
		"version": {version},
	})
	c.ReplyIfError(err)
	err = context.Call(&ret, cnc.Deploy, url.Values {
		"version": {version},
	})
	c.ReplyIfError(err)
	err = context.Call(nil, admin.Status, url.Values {
		"domains": domains,
		"status": {"3"},
	})
	c.ReplyIfError(err)
	err = context.Call(&ret, cnc.WaitForFinish, url.Values {
		"version": {version},
	})
	c.ReplyIfError(err)

	for _, dm := range domains {
		go func() {
			err = context.Call(&ret, CheckAndPass, url.Values {
				"domain": {dm},
			})
			c.Error(err)
		}()
	}
}
