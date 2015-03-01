package main

import (
	"github.com/layeh/gumble/gumble"
	"regexp"
	"strings"
)

type Expression struct {
	Value       string
	Description string
	Action      func(string)
}
type Command struct {
	Value  string
	Usage  string
	Action func([]string, *gumble.User)
}
type MessageParser struct {
	commands    map[string]*Command
	expressions []*Expression

	usagetemplate string
}

func (m *MessageParser) New() {
	m.commands = make(map[string]*Command)
}
func (m *MessageParser) LoadUsageTemplate() {
}
func (m *MessageParser) Usage() string {
	return "TODO"
}

func (m *MessageParser) RegisterCommand(value string, usage string, action func(args []string, sender *gumble.User)) {
	m.commands[value] = &Command{value, usage, action}
}
func (m *MessageParser) RegisterExpression(exp string, description string, action func(string)) error {
	_, err := regexp.Compile(exp)
	if err != nil {
		return err
	}
	m.expressions = append(m.expressions, &Expression{exp, description, action})
	return nil
}
func (m *MessageParser) Parse(message string, sender *gumble.User) {
	argv := strings.Split(message, " ")

	for _, expression := range m.expressions {
		exp, _ := regexp.Compile(expression.Value)
		for _, chunk := range argv {
			if exp.MatchString(chunk) {
				match := exp.FindString(chunk)
				expression.Action(match)
			}
		}
	}
	for commandkey, command := range m.commands {
		if strings.Index(commandkey, argv[0]) == 0 {
			command.Action(argv[1:len(argv)], sender)
		}
	}
}
