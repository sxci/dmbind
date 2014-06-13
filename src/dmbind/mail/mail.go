package mail

import (
	"strings"
	"net/smtp"
	"encoding/base64"
)

type MailProfile struct {
	User string
	Pswd string
	Port string
	Server string
}

func SendMail(mp MailProfile, to, cc []string, from, title, body string) (err error) {
	auth := smtp.PlainAuth("", mp.User, mp.Pswd, mp.Server)
	body = strings.TrimSpace(body)

	header := make(map[string]string)
	header["From"] = from
	header["To"] = strings.Join(to, ";")
	if len(cc) > 0 {
		header["Cc"] = strings.Join(cc, ";")
	}
	title = base64.StdEncoding.EncodeToString([]byte(title))
	header["Subject"] = "=?UTF-8?B?" + title + "?="
	header["MIME-Version"] = "1.0"
	header["Content-Type"] = `text/plain; charset="utf-8"`
	header["Content-Transfer-Encoding"] = "base64"
	if strings.HasPrefix(body, "<") {
		header["Content-Type"] = `text/html; charset="utf-8"`
	}

	message := ""
	for k, v := range header {
		message += k + ": " + v + "\r\n"
	}
	message += "\r\n" + base64.StdEncoding.EncodeToString([]byte(body))

	addr := mp.Server + ":" + mp.Port
	to = append(to, cc...)
	err = smtp.SendMail(addr, auth, from, to, []byte(message))
	return
}
