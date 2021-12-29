package command

import (
	"fmt"
	"os/exec"

	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/matcher"
	"github.com/innogames/slack-bot/v2/bot/msg"
)

// NewIPCommand is able to allow any ip
func NewIPCommand(base bot.BaseCommand) bot.Command {
	return &IPCommand{base}
}

type IPCommand struct {
	bot.BaseCommand
}

func (c *IPCommand) GetMatcher() matcher.Matcher {
	return matcher.NewRegexpMatcher(`ip +(?P<op>\w+) +(?P<ip>\d+\.\d+\.\d+\.\d+)`, c.execIP)
}

func run_script(args []string) string {
	cmd, err := exec.Command("/bin/sh", args...).Output()
	if err != nil {
		fmt.Printf("error %s", err)
	}
	output := string(cmd)
	return output
}

func (c *IPCommand) execIP(match matcher.Result, message msg.Message) {
	op := match.GetString("op")
	ip := match.GetString("ip")
	out := run_script([]string{"/root/myvim/slackbot/allow_ip.sh", op, ip})
	if op == "check" {
		if out != "" {
			out = fmt.Sprintf("%s is in blacklist", ip)
		} else {
			out = fmt.Sprintf("%s is not in blacklist", ip)
		}
		c.SlackClient.SendMessage(
			message,
			out,
		)
		return
	}
	c.SlackClient.SendMessage(
		message,
		fmt.Sprintf("%s %s done", op, ip),
	)
}

func (c *IPCommand) GetHelp() []bot.Help {
	return []bot.Help{
		{
			Command:     "ip <op> <ip>",
			Description: "allow/check/ban ip to connect to the internet",
			Examples: []string{
				"ip allow 192.168.18.44",
				"ip check 192.168.18.44",
			},
		},
	}
}
