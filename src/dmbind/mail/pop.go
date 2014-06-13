package mail

import (
	"os"
	"time"
	"errors"
	"strings"
	"strconv"
	"net/mail"
	"io/ioutil"
)

import (
	"dmbind/pop3"
)

var (
	storeFile = "mailTmp.txt"
	layout = "2006-01-02 15:04:05 MST"
	defaultDate = time.Date(1970, time.January, 1, 0, 0, 0, 0, time.Now().Location())
)

type MailServer struct {
	pop *pop3.Pop3
	prev time.Time
}

func NewMailServer(user, pass, mailHost string) (s *MailServer, err error) {
	s = &MailServer{}
	if err != nil { return }
	s.pop, err = pop3.NewPop(mailHost)
	if err != nil { return }
	s.prev, err = LoadDate()
	if err != nil { return }
	s.pop.Login(user, pass)
	return
}

type Mail struct {
	Header  mail.Header
	Subject string
	Body    string
	All     [][]byte
	Date    time.Time
	Mid     string
}

func (s *MailServer) Retrieve(idx int) (m Mail, err error) {
	h, body, err := s.pop.RetrieveOne(idx)
	if err != nil { return }
	date, err := h.Date()
	if err != nil {
		date = defaultDate
		err = nil
	}
	if len(body) == 0 {
		err = errors.New("miss body")
		return
	}
	m = Mail {h, h.Get("Subject"), string(body[0]), body, date, h.Get("X-Qq-Mid")}
	return
}

func (s *MailServer) FindMail(find string, after time.Time) (rets []Mail, err error) {
	rets = make([]Mail, 1024)
	length := 0
	count, err := s.pop.Count()
	if err != nil { return }
	t := s.prev
	for i:=count; i>0; i-- {
		m, err := s.Retrieve(i)
		if err != nil { continue }
		if m.Date.After(t) { t = m.Date }
		if m.Date.Before(after) || m.Date.Equal(after) { break }
		if ! strings.Contains(m.Subject, find) { continue }
		isSame := false
		for _, r := range rets {
			if r.Mid == m.Mid {
				isSame = true
				break
			}
		}
		if isSame { continue }
		rets[length] = m
		length ++
	}
	rets = rets[:length]
	if t.After(s.prev) {
		s.prev = t
		SaveDate(s.prev)
	}
	return
}

func RestoreData() (data string, err error) {
	r, err := ioutil.ReadFile(storeFile)
	if err != nil { return }
	data = strings.TrimSpace(string(r))
	return
}

func SaveDate(t time.Time) (err error) {
	return StoreData(strconv.FormatInt(t.Unix(), 10))
}

func LoadDate() (t time.Time, err error) {
	data, err := RestoreData()
	if err != nil {
		t, err = time.Parse(layout, "2013-09-09 00:00:00 CST")
		return
	}
	stamp, err := strconv.ParseInt(data, 10, 0)
	if err != nil { return }
	t = time.Unix(stamp, 0)
	return
}

func StoreData(data string) (err error) {
	err = ioutil.WriteFile(storeFile, []byte(data), os.ModePerm)
	return
}

func (s *MailServer) Close() (err error) {
	return s.pop.Close()
}
