package imap

import (
	"net"
	"bufio"
	"errors"
	"strconv"
	"strings"
)

type IMAP struct {
	reader *bufio.Reader
	conn net.Conn
	total int
	welcome string
}

func NewIMAP(host string) (p *IMAP, err error) {
	p = &IMAP{}
	conn, err := net.Dial("tcp", host)
	if err != nil { return }
	p.conn = conn
	p.reader = bufio.NewReader(conn)
	p.welcome, err = p.Readline()
	return
}

func (i *IMAP) Readline() (line string, err error) {
	return i.Read('\n')
}

func (i *IMAP) Read(char byte) (ret string, err error) {
	status, err := i.reader.ReadString(char)
	if err != nil { return }
	ret = string(status)
	return
}

func (i *IMAP) Send(data string) (err error) {
	i.conn.Write([]byte(data + "\n"))
	return
}

func (i *IMAP) Close() (err error) {
	if i.conn == nil { return }
	return i.conn.Close()
}

func (i *IMAP) Login(user, pswd string) (err error) {
	err = i.Send("a001 login " + user + " " + pswd)
	if err != nil { return }
	ret, err := i.Readline()
	if err != nil { return }
	if strings.Contains(ret, "ok") { return }
	err = errors.New(ret)
	return
}

type SelectRet struct {
	Total int
}

func (i *IMAP) Select(mailbox string) (ret SelectRet, err error) {
	rets, err := i.Call("a002 select " + mailbox)
	if err != nil { return }
	total, err := strconv.Atoi(strings.Split(rets[0], " ")[0])
	if err != nil { return }
	ret = SelectRet {
		Total: total,
	}
	return
}

func (i *IMAP) Fetch(start, end int, field string) (data []string, err error) {
	if start > end { return }
	cmd := "a002 fetch "+strconv.Itoa(start)+":"+strconv.Itoa(end)+" "+field
	err = i.Send(cmd)
	if err != nil { return }
	ret, err := i.ReadString("completed\r\n")
	if err != nil { return }
	fetchs := strings.Split(ret, "\r\n* ")
	cnc, length := make([]string, len(fetchs)), 0
	for _, f := range fetchs {
		if strings.Contains(f, "572R5a6/5Yqg6YCf5pyN5Yqh5rWL6K+V5oyH5Y2XLeiHtOS4g+eJm+S6k") {
			if idx := strings.Index(f, " "); idx > 0 {
				cnc[length] = f[:idx]
				length++
			}
		}
	}
	cnc = cnc[:length]
	data = cnc
	return
}

func (i *IMAP) GetContent(idx string) (data string, err error) {
	err = i.Send("a002 fetch " + idx + " RFC822.TEXT")
	if err != nil { return }
	ret, err := i.ReadString("completed\r\n")
	if err != nil { return }
	println(ret)
	return
}

func (i *IMAP) CheckIsUpdate() (start, end int, err error) {
	ret, err := i.Select("inbox")
	if err != nil { return }
	total := ret.Total
	if i.total >= total { return }
	start, end = ret.Total+1, total
	return
}

func (i *IMAP) Call(command string) (rets []string, err error) {
	err = i.Send(command)
	if err != nil { return }
	lines := ""
	var ret string
	for {
		ret, err = i.Readline()
		if err != nil { return }
		ret = strings.TrimSpace(ret)
		if len(ret) > 0 && ret[0] == '*' {
			lines += ret[2:] + "\n"
		} else {
			lines += ret + "\n"
			if strings.Count(lines, "(") == strings.Count(lines, ")") {
				break
			}
		}
	}
	rets = strings.Split(lines, "\n")
	return
}

func (i *IMAP) ReadString(str string) (ret string, err error) {
	for {
		tmp := ""
		tmp, err = i.Read(str[len(str)-1])
		if err != nil { return }
		ret += tmp
		if ret[len(ret)-len(str):] == str {
			break
		}
	}
	return
}
