package main

import (
	"bytes"
	"fmt"
	"github.com/layeh/gumble/gumble"
	"html/template"
	"regexp"
	"strings"
)

const (
	usagehtml = "templates/usage.html"
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
	Commands    map[string]*Command
	Expressions []*Expression

	usagetemplate *template.Template
}

func (m *MessageParser) New() {
	m.Commands = make(map[string]*Command)
	var err error
	m.usagetemplate, err = template.ParseFiles(usagehtml)
	if err != nil {
		panic(err)
	}
}
func (m *MessageParser) Usage() string {
	var buffer bytes.Buffer
	err := m.usagetemplate.Execute(&buffer, m)
	if err != nil {
		panic(err)
	}
	return buffer.String()
}

func (m *MessageParser) RegisterCommand(value string, usage string, action func(args []string, sender *gumble.User)) {
	m.Commands[value] = &Command{value, usage, action}
}
func (m *MessageParser) RegisterExpression(exp string, description string, action func(string)) error {
	_, err := regexp.Compile(exp)
	if err != nil {
		return err
	}
	m.Expressions = append(m.Expressions, &Expression{exp, description, action})
	return nil
}
func (m *MessageParser) Parse(message string, sender *gumble.User) {
	argv := strings.Split(message, " ")

	fmt.Println(argv)
	for _, expression := range m.Expressions {
		exp, _ := regexp.Compile(expression.Value)
		for _, chunk := range argv {
			if exp.MatchString(chunk) {
				match := exp.FindString(chunk)
				expression.Action(match)
			}
		}
	}
	for commandkey, command := range m.Commands {
		if commandkey == argv[0] {
			command.Action(argv[1:len(argv)], sender)
		}
	}
}
func SendMumbleMessage(message string, client *gumble.Client, channel *gumble.Channel) {
	textmessage := gumble.TextMessage{
		Channels: []*gumble.Channel{
			channel,
		},
		Message: message,
	}
	client.Send(&textmessage)
}
func SendMumbleMessageTo(user *gumble.User, message string, client *gumble.Client) {
	textmessage := gumble.TextMessage{
		Users: []*gumble.User{
			user,
		},
		Message: message,
	}
	client.Send(&textmessage)
}
