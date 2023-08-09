package command

import (
	"fmt"
	"os/exec"

	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/matcher"
	"github.com/innogames/slack-bot/v2/bot/msg"
)

// NewShellCommand is able to allow any ip
func NewShellCommand(base bot.BaseCommand) bot.Command {
	return &ShellCommand{base}
}

type ShellCommand struct {
	bot.BaseCommand
}

func (c *ShellCommand) GetMatcher() matcher.Matcher {
	return matcher.NewRegexpMatcher(`(?P<cmd>file|fq) +(?P<op>[\w|\-|\d|\.]+) +(?P<arg1>[\w|\-|\d|\.|:|\/]+)`, c.execCmd)
}

func cmd_run_script(args []string) string {
	cmd, err := exec.Command("/bin/sh", args...).Output()
	var output string
	if err != nil {
		output = fmt.Sprintf("ERROR:\n%s", err)
	} else {
		//output = string(cmd)
		output = fmt.Sprintf("OK:\n%s", cmd)
	}
	return output
}

func (c *ShellCommand) execCmd(match matcher.Result, message msg.Message) {
	cmd := match.GetString("cmd")
	op := match.GetString("op")
	arg1 := match.GetString("arg1")
	cmd_file := fmt.Sprintf("/root/myvim/slackbot/%s.sh", cmd)
	out := cmd_run_script([]string{cmd_file, op, arg1})
	c.SlackClient.SendMessage(
		message,
		out,
	)
}

func (c *ShellCommand) GetHelp() []bot.Help {
	return []bot.Help{
		{
			Command:     "<cmd> <op> <arg1>",
			Description: "execute any shell script",
		},
	}
}
