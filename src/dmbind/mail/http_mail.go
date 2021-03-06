package mail

import (
	"dmbind/lib/context"
)

var profiles map[string]MailProfile

func SetupProfile(profile map[string]MailProfile) {
	profiles = profile
}

func Send(c *context.Context) {
	to := c.MustStrings("to")
	to = toAddress(to)

	cc := toAddress(c.Strings("cc", nil))
	subject, content := c.MustString("subject"), c.MustString("content")
	profileName := c.String("profile", "noreply")
	profile, ok := profiles[profileName]
	if !ok {
		c.ReplyErrorInfo("invalid profile name " + profileName)
		return
	}
	err := SendMail(profile, to, cc, profile.User, subject, content)
	c.ReplyIfError(err)
	c.Info("sending mail to", to)
	c.ReplyObj(true)
}

func toAddress(as []string) (bs []string) {
	if as == nil {
		return
	}
	bs = make([]string, len(as))
	for idx, val := range as {
		if p, ok := profiles[val]; ok {
			bs[idx] = p.User
		} else {
			bs[idx] = val
		}
	}
	return
}
