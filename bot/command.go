package bot

import (
	"reflect"
	"sort"
	"strings"
	"sync"

	"github.com/innogames/slack-bot/v2/bot/matcher"
	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/innogames/slack-bot/v2/bot/util"
	"github.com/innogames/slack-bot/v2/client"
)

var lock sync.Mutex

// Command is the main command struct which needs to provide the matcher and the actual executed action
type Command interface {
	GetMatcher() matcher.Matcher
}

// BaseCommand is base struct which is handy for all commands, as a direct Slack communication is possible
type BaseCommand struct {
	client.SlackClient
}

// Conditional checks if the command should be activated. E.g. is dependencies are not present or it's disabled in the config
type Conditional interface {
	IsEnabled() bool
}

// Runnable indicates that the command executes a go function
type Runnable interface {
	RunAsync()
}

// HelpProvider can be provided by a command to add information within "help" command
type HelpProvider interface {
	// GetHelp each command should provide information, like a description or examples
	GetHelp() []Help
}

// Commands is a wrapper of a list of commands. Only the first matched command will be executed
type Commands struct {
	commands     []Command
	matcher      []matcher.Matcher // precompiled matcher objects
	matcherNames map[int]string    // precompiled mapping from matcher -> command name
	compiled     bool
}

// GetHelp returns the help for ALL included commands
func (c *Commands) GetHelp() []Help {
	help := make([]Help, 0)

	for _, command := range c.commands {
		if helpCommand, ok := command.(HelpProvider); ok {
			help = append(help, helpCommand.GetHelp()...)
		}
	}

	return help
}

// Run executes the first matched command and return true in case one command matched
func (c *Commands) Run(message msg.Message) bool {
	matched, _ := c.RunWithName(message)

	return matched
}

// RunWithName executes the first matched command and return the command name if there was a match
func (c *Commands) RunWithName(message msg.Message) (bool, string) {
	c.compile()

	for i, command := range c.matcher {
		run, match := command.Match(message)
		if match != nil {
			// this is is needed for ConditionMatcher: runner gets already executed in the matcher itself!
			if run != nil {
				run(match, message)
			}

			// only the first command is executed -> abort here
			return true, c.matcherNames[i]
		}
	}

	return false, ""
}

// AddCommand registers a command to the command list
func (c *Commands) AddCommand(commands ...Command) {
	for _, command := range commands {
		if command == nil {
			continue
		}

		if condition, ok := command.(Conditional); ok {
			if !condition.IsEnabled() {
				// command is disabled!
				continue
			}
		}

		// register template function defined in commands
		if provider, ok := command.(util.TemplateFunctionProvider); ok {
			util.RegisterFunctions(provider.GetTemplateFunction())
		}

		c.commands = append(c.commands, command)
	}

	c.compiled = false
}

// Merge two list of commands
func (c *Commands) Merge(commands Commands) {
	c.AddCommand(commands.commands...)
}

// Count the registered/valid commands
func (c *Commands) Count() int {
	c.compile()

	return len(c.commands)
}

func (c *Commands) GetCommandNames() []string {
	c.compile()

	names := make([]string, 0, len(c.matcherNames))
	for _, name := range c.matcherNames {
		names = append(names, name)
	}

	sort.Strings(names)

	return names
}

func (c *Commands) compile() {
	if c.compiled {
		return
	}

	// make sure only one process is creating the compiled list
	lock.Lock()
	defer lock.Unlock()

	if !c.compiled {
		c.matcher = make([]matcher.Matcher, len(c.commands))
		c.matcherNames = make(map[int]string, len(c.commands))

		for i, command := range c.commands {
			commandName := getCommandName(command)
			commandMatcher := command.GetMatcher()

			c.matcher[i] = commandMatcher
			c.matcherNames[i] = commandName
		}
		c.compiled = true
	}
}

func getCommandName(command Command) string {
	t := reflect.TypeOf(command)
	name := t.String()
	name = strings.ReplaceAll(name, "*", "")
	name = strings.TrimSuffix(name, "Command")

	return name
}
