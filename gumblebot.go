package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"html/template"
	"strings"
	"github.com/layeh/gumble/gumble_ffmpeg"
	"github.com/layeh/gumble/gumble"
	"github.com/layeh/gumble/gumbleutil"
	"encoding/gob"
	"flag"
)
const helpTemplate = `
<b>Gumblebot Help</b>
<ul>
    <li>
       Soundboard Commands:
       <ul>
		<b><li> sbon to enable soundboard, sboff to disable soundboard </li></b>
		{{range $key, $value := .}}
		<li>
            {{$key}}
		</li>
        {{end}}
		</ul>
    </li>
</ul>`

const datafile = "data"

type SoundboardUser struct {
	SoundboardEnabled bool
	SoundboardVolume float64
	WelcomeSound string
}

var soundboard_users map[string]SoundboardUser

func send_usage(client *gumble.Client, soundboard map[string]string) {
	var buffer bytes.Buffer

	outTemplate, err := template.New("help").Parse(helpTemplate)
	if err != nil {
		panic(err)
	}

	err = outTemplate.Execute(&buffer, soundboard )
	if err != nil {
		panic(err);
	}
	message := gumble.TextMessage{
		Channels: []*gumble.Channel{
			client.Self.Channel,
		},
		Message: buffer.String(),
	}
	client.Send(&message)
}
func decode_users() {
	iobuffer, err := ioutil.ReadFile(datafile)
	if err != nil {
		fmt.Println(err)
		return
	}
	buffer := bytes.NewBuffer(iobuffer)
	dec := gob.NewDecoder(buffer)

	err = dec.Decode(&soundboard_users)
	if err != nil {
		panic(err)
	}
}
func encode_users() {
	var buffer bytes.Buffer
	enc := gob.NewEncoder(&buffer)
	err := enc.Encode(soundboard_users)
	if err != nil {
		panic(err)
	}
	err = ioutil.WriteFile(datafile, buffer.Bytes(), 0755)
	if err != nil {
		panic(err)
	}
}
func update_users(client *gumble.Client) {
	for _, user := range client.Users {
		if _, ok := soundboard_users[user.Name]; ok {
			continue
		}
		soundboard_users[user.Name] = SoundboardUser{false, 0.5, ""}
		encode_users()
	}
}
func init() {
	soundboard_users = make(map[string]SoundboardUser)

	decode_users()
}


func main() {
	var sounds_dir = flag.String("sounds", "sounds", "directory where soundboard files are located")
	var volume = flag.Float64("volume", 0.25, "soundboard volume from 0 to 1")
	var stream *gumble_ffmpeg.Stream
	soundboard := make(map[string]string)

	// Populate soundboard with files in $pwd/sounds/
	files, err := ioutil.ReadDir(*sounds_dir)
	if err != nil  {
		fmt.Println(err)
		return
	}
	for _, f := range files {
		key := f.Name()
		if filepath.Ext(key) == ".mp3" || filepath.Ext(key) == ".ogg" || filepath.Ext(key) == ".wav" {
			value, _ := filepath.Abs(filepath.Join(*sounds_dir, key))
			soundboard[key] = value
		}
	}


	var gumbleclient *gumble.Client

	gumbleutil.Main(func(client *gumble.Client) {
		stream, _ = gumble_ffmpeg.New(client)
		stream.Volume = float32(*volume)
		client.Attach(gumbleutil.AutoBitrate)
		gumbleclient = client
	}, gumbleutil.Listener{
		Connect: func(e *gumble.ConnectEvent) {
			fmt.Printf("Connected!\n")
			update_users(gumbleclient)
		},
		TextMessage: func(e *gumble.TextMessageEvent) {
			if e.Sender == nil {
				return
			}
			if e.Message == "stop" {
				stream.Stop()
			}
			if e.Message == "help" {
				go send_usage(gumbleclient, soundboard)
			}
			if e.Message == "sboff" {
				u := soundboard_users[e.Sender.Name]
				u.SoundboardEnabled = false
				soundboard_users[e.Sender.Name] = u
				encode_users()
			}
			if e.Message == "sbon" {
				u := soundboard_users[e.Sender.Name]
				u.SoundboardEnabled = true
				soundboard_users[e.Sender.Name] = u
				encode_users()
			}
			for key, value := range soundboard {
				if strings.Index(key, e.Message) == 0 {
					vtarget := &gumble.VoiceTarget{}
					vtarget.ID = 1
					for username, sb := range soundboard_users {
						userstruct := gumbleclient.Users.Find(username)
						if userstruct != nil && sb.SoundboardEnabled == true {
							vtarget.AddUser(userstruct)
						}
					}
					stream.Stop()
					gumbleclient.Send(vtarget)
					gumbleclient.VoiceTarget = vtarget
					stream.Play(value)
				}
			}
		},
		UserChange: func(e *gumble.UserChangeEvent) {
			update_users(gumbleclient)
		},
	})

}

