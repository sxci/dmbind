package pub

import (
	"net/http"
)
type rsClient struct {
	Conn Client
}

func New(mac *Mac) rsClient {
	t := NewTransport(mac, nil)
	client := &http.Client{Transport: t}
	return rsClient{Client{client}}
}

func NewEx(t http.RoundTripper) rsClient {
	client := &http.Client{Transport: t}
	return rsClient{Client{client}}
}
