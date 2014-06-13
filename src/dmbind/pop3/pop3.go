package pop3

import (
	"io"
	"fmt"
	"net"
	"bufio"
	"errors"
	"strings"
	"strconv"
	"net/url"
	"net/mail"
	"io/ioutil"
	"mime/multipart"
	"encoding/base64"
)


var (
	errorResultNotMatch = errors.New("result not match")
)

type Pop3 struct {
	reader *bufio.Reader
	conn net.Conn
	welcome string
}

func NewPop(host string) (p *Pop3, err error) {
	p = &Pop3{}
	conn, err := net.Dial("tcp", host)
	if err != nil { return }
	p.conn = conn
	p.reader = bufio.NewReader(conn)
	p.welcome, err = p.Readline()
	return
}

func (p Pop3) Count() (count int, err error) {
	err = p.Send("stat")
	if err != nil { return }
	ret, err := p.Readline()
	if err != nil { return }
	rets := strings.Split(ret, " ")
	if len(rets) <= 1 {
		err = errors.New("result not except")
		return
	}
	count, err = strconv.Atoi(rets[1])
	return
}

func (p Pop3) Login(user, pasw string) (err error) {
	err = p.Except("user " + user, "+OK\r\n")
	if err != nil { return }
	err = p.Except("pass " + pasw, "+OK\r\n")
	return
}

func (p Pop3) Send(data string) (err error) {
	p.conn.Write([]byte(data + "\n"))
	return
}

func (p Pop3) Read(char byte) (ret string, err error) {
	status, err := p.reader.ReadString(char)
	if err != nil { return }
	ret = string(status)
	return
}

func (p Pop3) Readline() (ret string, err error) {
	return p.Read('\n')
}

func (p Pop3) ReadString(str string) (ret string, err error) {
	for {
		tmp := ""
		tmp, err = p.Read(str[len(str)-1])
		if err != nil { return }
		ret += tmp
		if ret[len(ret)-len(str):] == str {
			break
		}
	}
	return
}

func (p Pop3) Except(send, except string) (err error) {
	err = p.Send(send)
	if err != nil { return }
	ret, err := p.Readline()
	if err != nil { return }
	if ret != except {
		fmt.Println("result not except", []byte(ret), []byte(except))
		err = errorResultNotMatch
	}
	return
}

func (p Pop3) Close() (err error) {
	if p.conn == nil {
		return
	}
	return p.conn.Close()
}

func (p Pop3) DecodeBody(stream io.Reader) (h mail.Header, data [][]byte, err error) {
	r, err := mail.ReadMessage(stream)
	if err != nil { return }
	h = r.Header
	if subject := DecodeSubject(h.Get("Subject")); subject != "" {
		h["Subject"] = []string {subject}
	}
	contentType := r.Header.Get("Content-Type")
	if contentType == "" {
		var d []byte
		d, err = ioutil.ReadAll(r.Body)
		if err != nil { return }
		data = [][]byte {d}
		return
	}

	boundary := ""
	encoding := "base64"
	if idx := strings.Index(contentType, "boundary"); idx > 0 {
		boundary = contentType[idx+9: ]
		if idx := strings.Index(boundary, `"`); idx >= 0 {
			boundary = boundary[idx+1:]
		}
		if idx := strings.Index(boundary, `"`); idx > 0 {
			boundary = boundary[:idx]
		}
		
		rd := multipart.NewReader(r.Body, boundary)
		var part *multipart.Part
		var d []byte
		for {
			part, err = rd.NextPart()
			if err == io.EOF {
				err = nil
				break
			}
			if err != nil { return }
			d, err = ioutil.ReadAll(part)
			if err != nil { return }
			encoding = part.Header.Get("Content-Transfer-Encoding")
			d = decodeString(string(d), encoding)
			data = append(data, d)
		}
	} else {
		var d []byte
		d, err = ioutil.ReadAll(r.Body)
		if err != nil { return }
		data = [][]byte { decodeString(string(d), encoding) }
	}
	return
}

func decodeString(s, encoding string) (ret []byte) {
	var err error
	ret = []byte(s)
	switch encoding {
	case "base64":
		ret, err = base64.StdEncoding.DecodeString(s)
		if err != nil { ret = []byte(s) }
	}
	return
}

func (p Pop3) RetrieveOne(idx int) (h mail.Header, data [][]byte, err error) {
	ret, err := p.Retr(idx)
	if err != nil { return }
	return p.DecodeBody(strings.NewReader(ret))
}

func (p Pop3) Retr(idx int) (ret string, err error) {
	err = p.Send("retr " + strconv.Itoa(idx))
	if err != nil { return }
	ret, err = p.Readline()
	if err != nil { return }
	if ret != "+OK\r\n" {
		err = errors.New(ret)
		return
	}
	ret, err = p.ReadString("\r\n.\r\n")
	if err != nil { return }
	ret = ret[: len(ret)-5]
	return
}

func DecodeSubject(subject string) (ret string) {
	subjectSlice := strings.Split(subject, " ")
	for _, sb := range subjectSlice {
		if ! strings.HasPrefix(sb, "=?") || ! strings.HasSuffix(sb, "?=") {
			ret += sb
			continue
		}

		decoded := ""
		data := strings.Split(sb, "?")
		// encoding := strings.ToLower(data[1])
		switch strings.ToUpper(data[2]) {
			case "B":
				ret1, err := base64.StdEncoding.DecodeString(data[3])
				if err != nil { continue }
				decoded = string(ret1)
			case "Q":
				var err error
				tmp := strings.Replace(data[3], "_", "%20", -1)
				tmp = strings.Replace(tmp, "=", "%", -1)
				tmp, err = url.QueryUnescape(tmp)
				if err != nil {
					continue
				}
				decoded = string(tmp)
		}
		// if encoding != "utf-8" {
			// var err error
			// decoded, err = iconv.ConvertString(decoded, encoding, "utf-8")
			// if err != nil { continue }
		// }
		ret += decoded
	}
	return
}

func (p Pop3) List() (rets []string, err error) {
	err = p.Send("list")
	if err != nil { return }
	ret, err := p.Read('.')
	p.Readline()
	if err != nil { return }
	lineSp := strings.Split(ret, "\r\n")
	length := 0
	rets = make([]string, len(lineSp))
	for _, r := range lineSp {
		sp := strings.Split(r, " ")
		if len(sp) != 2 { continue }
		
		rets[length] = sp[1]
		length++
	}
	rets = rets[:length]
	return
}
