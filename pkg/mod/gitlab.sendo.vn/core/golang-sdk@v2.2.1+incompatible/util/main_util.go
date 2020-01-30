package util

import (
	"fmt"
	"os"
	"time"
)

type SimpleCommand struct {
	Name string
	Desc string
	Func func(args []string)
}

type SimpleMain struct {
	cmds []*SimpleCommand
}

func (m *SimpleMain) Add(cmd *SimpleCommand) {
	m.cmds = append(m.cmds, cmd)
}

func (m *SimpleMain) Execute() {
	if len(os.Args) < 2 {
		m.usage()
	}
	// split command & sub-args
	cmdName := os.Args[1]
	args := []string{}
	if len(os.Args) > 2 {
		args = os.Args[2:]
	}

	for _, c := range m.cmds {
		if c.Name == cmdName {
			c.Func(args)
			goto SHUTDOWN
		}
	}

	m.usage()

SHUTDOWN:
}

func (m *SimpleMain) usage() {
	s := os.Args[0] + ":\n"

	if inf := GetBuildInfo(); inf != nil {
		s += fmt.Sprintf(`
RELEASE INFORMATION:
  Date     : %s
  Branch   : %s
  REVISION : %s

`,
			inf.GetDate().Format(time.RFC1123),
			inf.GetBranch(),
			inf.Revision,
		)
	}

	s += "Usage:\n"

	for _, c := range m.cmds {
		s += fmt.Sprintf("  %-10s %s\n", c.Name, c.Desc)
	}

	fmt.Println(s)
	os.Exit(2)
}
