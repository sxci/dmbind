package domain

import (
	"strings"
	"os/exec"
	"dmbind/lib/context"
)

func shell(cmd string) ([]byte, error) {
	ret, err := exec.Command("bash", "-c", cmd).Output()
	if err != nil { return nil, err }
	return ret, nil
}

func TxWs(c *context.Context) {
	ret, err := shell("cat dmbind.err| grep 'DnspodCName' | grep 'success'| awk '{print $(NF-2)}'")
	c.ReplyIfError(err)
	buckets := strings.Split(string(ret), ".u\n")
	c.ReplyObj(buckets)
}
