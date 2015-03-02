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
	GumblebotRoot      = "root"
	GumblebotModerator = "moderator"
	GumblebotUser      = "user"

	permissiondenied = "I'm sorry dave, I can't do that."
	whoistemplate    = `
		<br></br><b> Whois {{ .Name }} </b>
		<ul>
			<li>{{.AccessLevel}} </li>
		<ul>`
)

type AdminUser struct {
	UserName        string
	MoveAllowed     bool
	KickAllowed     bool
	BanAllowed      bool
	RegisterAllowed bool
	AccessLevel     string
}

type MumbleAdmin struct {
	Users  map[string]*AdminUser
	Client *gumble.Client
}
type WhoisContext struct {
	Name        string
	AccessLevel string
}

func (m *MumbleAdmin) Attach(client *gumble.Client) {
	m.Client = client
}
func (m *MumbleAdmin) LoadAdminData(datapath string) {
	m.Users = make(map[string]*AdminUser)
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
func (m *MumbleAdmin) RegisterUser(user string, accesslevel string) {
	switch accesslevel {
	case GumblebotRoot:
		m.Users[user] = &AdminUser{UserName: user, MoveAllowed: true, KickAllowed: true, BanAllowed: true, RegisterAllowed: true, AccessLevel: accesslevel}
	case GumblebotModerator:
		m.Users[user] = &AdminUser{UserName: user, MoveAllowed: true, KickAllowed: true, BanAllowed: true, RegisterAllowed: false, AccessLevel: accesslevel}
	case GumblebotUser:
		m.Users[user] = &AdminUser{UserName: user, MoveAllowed: false, KickAllowed: false, BanAllowed: false, RegisterAllowed: false, AccessLevel: accesslevel}
	}
}
func (m *MumbleAdmin) search_mumble_users_substring(target string) *gumble.User {
	for _, user := range m.Client.Users {
		if strings.Index(strings.ToLower(user.Name), strings.ToLower(target)) == 0 {
			return user
		}
	}
	return nil
}
func (m *MumbleAdmin) Move(sender *gumble.User, channelsubstring string, users []string) {
	if user, ok := m.Users[sender.Name]; ok {
		if user.MoveAllowed != true {
			SendMumbleMessage(permissiondenied, m.Client, m.Client.Self.Channel)
			return
		}
		var targetChannel *gumble.Channel
		for _, channel := range m.Client.Channels {
			channelname := strings.ToLower(channel.Name)
			if strings.Index(channelname, strings.ToLower(channelsubstring)) == 0 {
				targetChannel = channel
			}
		}
		if targetChannel == nil {
			nochanerr := fmt.Sprintf("No such channel: %s", channelsubstring)
			SendMumbleMessage(nochanerr, m.Client, m.Client.Self.Channel)
			return
		}
		for _, targetusersubstring := range users {
			targetuser := m.search_mumble_users_substring(targetusersubstring)
			if targetuser == nil {
				nousererr := fmt.Sprintf("No such user: %s", targetusersubstring)
				SendMumbleMessage(nousererr, m.Client, m.Client.Self.Channel)
				return
			}
			targetuser.Move(targetChannel)
		}
	}
}
func (m *MumbleAdmin) Poke(sender *gumble.User, targetusername string) {
	targetuser := m.search_mumble_users_substring(targetusername)
	if targetuser != nil {
		SendMumbleMessageTo(targetuser, fmt.Sprintf("%s poked you!", sender.Name), m.Client)
	} else {
		SendMumbleMessage(fmt.Sprintf("%s not found!", targetusername), m.Client, m.Client.Self.Channel)
	}
}

func (m *MumbleAdmin) Whois(sender *gumble.User, targetusername string) {
	if _, ok := m.Users[sender.Name]; ok {
		targetuser := m.search_mumble_users_substring(targetusername)
		if targetuser == nil {
			// no such user, return!
			return
		}
		var accessLevel string
		if targetadmin, ok := m.Users[targetuser.Name]; ok {
			switch targetadmin.AccessLevel {
			case GumblebotRoot:
				accessLevel = "Root Administrator"
			case GumblebotModerator:
				accessLevel = "Mumble Moderator"
			case GumblebotUser:
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
		SendMumbleMessage(buffer.String(), m.Client, m.Client.Self.Channel)
	}
}
