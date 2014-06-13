package pub

import (
	"io"
	"net/http"
	"encoding/base64"
	"crypto/sha1"
	"crypto/hmac"
)

// ----------------------------------------------------------

type Mac struct {
	AccessKey string
	SecretKey []byte
}

func (mac *Mac) Sign(data []byte) (token string) {

	h := hmac.New(sha1.New, mac.SecretKey)
	h.Write(data)

	sign := base64.URLEncoding.EncodeToString(h.Sum(nil))
	return mac.AccessKey + ":" + sign[:27]
}

func (mac *Mac) SignWithData(b []byte) (token string) {

	blen := base64.URLEncoding.EncodedLen(len(b))

	key := mac.AccessKey
	nkey := len(key)
	ret := make([]byte, nkey+30+blen)

	base64.URLEncoding.Encode(ret[nkey+30:], b)

	h := hmac.New(sha1.New, mac.SecretKey)
	h.Write(ret[nkey+30:])
	digest := h.Sum(nil)

	copy(ret, key)
	ret[nkey] = ':'
	base64.URLEncoding.Encode(ret[nkey+1:], digest)
	ret[nkey+29] = ':'

	return string(ret)
}

func (mac *Mac) SignRequest(req *http.Request, incbody bool) (token string, err error) {

	h := hmac.New(sha1.New, mac.SecretKey)

	u := req.URL
	data := u.Path
	if u.RawQuery != "" {
		data += "?" + u.RawQuery
	}
	io.WriteString(h, data + "\n")
	sign := base64.URLEncoding.EncodeToString(h.Sum(nil))
	token = mac.AccessKey + ":" + sign
	return
}

func Sign(mac *Mac, data []byte) string {
	return mac.Sign(data)
}

func SignWithData(mac *Mac, data []byte) string {
	return mac.SignWithData(data)
}

// ---------------------------------------------------------------------------------------

type Transport struct {
	mac Mac
	transport http.RoundTripper
}

func incBody(req *http.Request) bool {

	if req.Body == nil {
		return false
	}
	if ct, ok := req.Header["Content-Type"]; ok {
		switch ct[0] {
		case "application/x-www-form-urlencoded":
			return true
		}
	}
	return false
}

func (t *Transport) RoundTrip(req *http.Request) (resp *http.Response, err error) {

	token, err := t.mac.SignRequest(req, incBody(req))
	if err != nil {
		return
	}
	req.Header.Set("Authorization", "QBox "+token)
	return t.transport.RoundTrip(req)
}

func NewTransport(mac *Mac, transport http.RoundTripper) *Transport {

	if transport == nil {
		transport = http.DefaultTransport
	}
	t := &Transport{transport: transport}
	t.mac = *mac
	return t
}

func NewClient(mac *Mac, transport http.RoundTripper) *http.Client {

	t := NewTransport(mac, transport)
	return &http.Client{Transport: t}
}

// ---------------------------------------------------------------------------------------


