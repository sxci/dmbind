package imap

import (
	"testing"
)

func TestImap(t *testing.T) {
	i, err := NewIMAP("imap.exmail.qq.com:143")
	if err != nil { t.Fatal(err) }
	err = i.Login("chenye@qiniu.com", "luoyecao8353364")
	if err != nil { t.Fatal(err) }
	_, _, err = i.CheckIsUpdate()
	ids, err := i.Fetch(9220, 9241, "body[header]")
	if err != nil { t.Fatal(err) }
	ret, err := i.GetContent(ids[0])
	if err != nil { t.Fatal(err) }
	t.Error(ret)
}
