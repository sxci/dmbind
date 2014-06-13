package domain

import (
	"errors"
	"strings"
)

func GetHost(domain string) (host string, err error) {
	for _, dm := range hostLib {
		if strings.HasSuffix(domain, dm) {
			host = domain
			if idx := strings.LastIndex(host[:len(host)-len(dm)], "."); idx > 0 {
				host = domain[idx+1:]
			}
			return
		}
	}
	err = errors.New("unknown domain")
	return
}


