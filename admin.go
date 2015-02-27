package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/layeh/gumble/gumble"
	"html/template"
	"io/ioutil"
	"strings"
)

const (
	GumblebotAdminRoot      = 2
	GumblebotAdminModerator = 1
	GumblebotAdminUser      = 0

	whoistemplate = `
		<b> Whois {{ .Name }} </b>
		<ul>
			<li>{{.AccessLevel}} </li>
		<ul>`
)

type AdminUser struct {
	UserName    string
	AccessLevel uint
}

type MumbleAdmin struct {
	Users map[string]AdminUser
}
type WhoisContext struct {
	Name        string
	AccessLevel string
}

func (m *MumbleAdmin) LoadAdminData(datapath string) {
	m.Users = make(map[string]AdminUser)
	iobuffer, err := ioutil.ReadFile(datapath)
	if err != nil {
		fmt.Println(err)
		return
	}
	buffer := bytes.NewBuffer(iobuffer)
	dec := gob.NewDecoder(buffer)
	err = dec.Decode(&m.Users)
	if err != nil {
		panic(err)
	}
}
func (m *MumbleAdmin) SaveAdminData(datapath string) {
	var buffer bytes.Buffer
	enc := gob.NewEncoder(&buffer)
	err := enc.Encode(m.Users)
	if err != nil {
		panic(err)
	}
	err = ioutil.WriteFile(datapath, buffer.Bytes(), 0755)
	if err != nil {
		panic(err)
	}
}
func (m *MumbleAdmin) RegisterUser(user string, accesslevel uint) {
	m.Users[user] = AdminUser{UserName: user, AccessLevel: accesslevel}
}
func (m *MumbleAdmin) search_users_substring(target string, client *gumble.Client) *gumble.User {
	for _, user := range client.Users {
		if strings.Index(user.Name, target) == 0 {
			return user
		}
	}
	return nil
}

func (m *MumbleAdmin) Whois(sender *gumble.User, targetusername string, client *gumble.Client) {
	if user, ok := m.Users[sender.Name]; ok {
		if user.AccessLevel >= GumblebotAdminUser {
			targetuser := m.search_users_substring(targetusername, client)
			if targetuser == nil {
				// no such user, return!
				return
			}
			var accessLevel string
			if targetadmin, ok := m.Users[targetuser.Name]; ok {
				switch targetadmin.AccessLevel {
				case GumblebotAdminRoot:
					accessLevel = "Root Administrator"
				case GumblebotAdminModerator:
					accessLevel = "Mumble Moderator"
				case GumblebotAdminUser:
					accessLevel = "Mumble User"
				}
			} else {
				accessLevel = "Mumble User"
			}

			var buffer bytes.Buffer
			template, err := template.New("whois").Parse(whoistemplate)
			if err != nil {
				panic(err)
			}
			err = template.Execute(&buffer, WhoisContext{targetuser.Name, accessLevel})

			if err != nil {
				panic(err)
			}
			message := gumble.TextMessage{
				Channels: []*gumble.Channel{
					client.Self.Channel,
				},
				Message: buffer.String(),
			}
			client.Send(&message)
		}
	}
}
