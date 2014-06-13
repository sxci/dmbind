package log

import (
	"os/exec"
	"dmbind/lib/context"
)

func View(c *context.Context) {
	logFile := c.String("filename", "dmbind.log")
	ret, err := exec.Command("tail", "-n", "100", logFile).CombinedOutput()
	c.ReplyIfError(err)
	c.ReplyText(string(ret))
}
