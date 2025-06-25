package cobra

import (
	"errors"
	"fmt"
	"os"
)

// Command represents a CLI command.
type Command struct {
	Use   string
	Short string
	RunE  func(cmd *Command, args []string) error
	Run   func(cmd *Command, args []string)

	commands []*Command
}

// AddCommand adds subcommands to this command.
func (c *Command) AddCommand(cmds ...*Command) {
	c.commands = append(c.commands, cmds...)
}

// Help returns a formatted help message for this command and any subcommands.
func (c *Command) Help() string {
	usage := fmt.Sprintf("Usage:\n  %s", c.Use)
	if len(c.commands) > 0 {
		usage += " [command]"
	}
	if c.Short != "" {
		usage += "\n\n" + c.Short
	}
	if len(c.commands) > 0 {
		usage += "\n\nAvailable Commands:\n"
		for _, sub := range c.commands {
			usage += fmt.Sprintf("  %s", sub.Use)
			if sub.Short != "" {
				usage += fmt.Sprintf("\t%s", sub.Short)
			}
			usage += "\n"
		}
	}
	return usage
}

// Execute runs the command using os.Args.
func (c *Command) Execute() error {
	args := os.Args[1:]
	return c.execute(args)
}

func (c *Command) execute(args []string) error {
	if len(args) == 0 {
		if c.RunE != nil {
			return c.RunE(c, nil)
		}
		if c.Run != nil {
			c.Run(c, nil)
			return nil
		}
		if len(c.commands) > 0 {
			fmt.Println(c.Help())
			return nil
		}
		return nil
	}

	switch args[0] {
	case "-h", "--help", "help":
		fmt.Println(c.Help())
		return nil
	}

	for _, sub := range c.commands {
		if args[0] == sub.Use {
			return sub.execute(args[1:])
		}
	}
	if c.RunE != nil {
		return c.RunE(c, args)
	}
	if c.Run != nil {
		c.Run(c, args)
		return nil
	}
	return errors.New("unknown command: " + args[0])
}
