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
type Soundboard struct {
	Users map[string]SoundboardUser
	sounds map[string]string
}
func (s *Soundboard) LoadSounds(path string) {
	s.sounds = make(map[string]string)
	files,err := ioutil.ReadDir(path)
	if err != nil {
		fmt.Println(err)
		return
	}
	for _, f := range files {
		key := f.Name()
		if filepath.Ext(key) == ".mp3" || filepath.Ext(key) == ".ogg" || filepath.Ext(key) == ".wav" {
			value, _ := filepath.Abs(filepath.Join(path, key))
			s.sounds[key] = value
		}
	}
}
func (s *Soundboard) LoadUsers(datapath string) {
	iobuffer, err := ioutil.ReadFile(datapath)
	if err != nil {
		fmt.Println(err)
		return
	}
	buffer := bytes.NewBuffer(iobuffer)
	dec := gob.NewDecoder(buffer)
	err = dec.Decode(&s.Users)
	if err != nil {
		panic(err)
	}
}
func (s *Soundboard) SaveUsers(datapath string) {
	var buffer bytes.Buffer
	enc := gob.NewEncoder(&buffer)
	err := enc.Encode(s.Users)
	if err != nil {
		panic(err)
	}
	err = ioutil.WriteFile(datapath, buffer.Bytes(), 0755)
	if err != nil {
		panic(err)
	}
}
func (s *Soundboard) UpdateUsers(client *gumble.Client) {
	for _, user := range client.Users {
		if _, ok := s.Users[user.Name]; ok {
			continue
		}
		s.Users[user.Name] = SoundboardUser{false, 0.5, ""}
	}
}
func (s *Soundboard) Play(client *gumble.Client, stream *gumble_ffmpeg.Stream, sound string)  {
	for key, value := range s.sounds {
		if strings.Index(key, sound) == 0 {
			vtarget := &gumble.VoiceTarget{}
			vtarget.ID = 1
			for username, sb := range s.Users {
				userstruct := client.Users.Find(username)
				if userstruct != nil && sb.SoundboardEnabled == true {
					vtarget.AddUser(userstruct)
				}
			}
			stream.Stop()
			client.Send(vtarget)
			client.VoiceTarget = vtarget
			stream.Play(value)
		}
	}
}

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

func main() {
	var sounds_dir = flag.String("sounds", "sounds", "directory where soundboard files are located")
	var volume = flag.Float64("volume", 0.25, "soundboard volume from 0 to 1")
	var stream *gumble_ffmpeg.Stream

	var soundboard Soundboard
	var gumbleclient *gumble.Client

	soundboard.LoadUsers(datafile)
	soundboard.LoadSounds(*sounds_dir)

	gumbleutil.Main(func(client *gumble.Client) {
		stream, _ = gumble_ffmpeg.New(client)
		stream.Volume = float32(*volume)
		client.Attach(gumbleutil.AutoBitrate)
		gumbleclient = client
	}, gumbleutil.Listener{
		Connect: func(e *gumble.ConnectEvent) {
			fmt.Printf("Connected!\n")
			soundboard.UpdateUsers(gumbleclient)
			soundboard.SaveUsers(datafile)
		},
		TextMessage: func(e *gumble.TextMessageEvent) {
			if e.Sender == nil {
				return
			}
			if e.Message == "stop" {
				stream.Stop()
			}
			if e.Message == "help" {
				go send_usage(gumbleclient, soundboard.sounds)
			}
			if e.Message == "sboff" {
				u := soundboard.Users[e.Sender.Name]
				u.SoundboardEnabled = false
				soundboard.Users[e.Sender.Name] = u
				soundboard.SaveUsers(datafile)
			}
			if e.Message == "sbon" {
				u := soundboard.Users[e.Sender.Name]
				u.SoundboardEnabled = true
				soundboard.Users[e.Sender.Name] = u
				soundboard.SaveUsers(datafile)
			}
			soundboard.Play(gumbleclient, stream, e.Message)
		},
		UserChange: func(e *gumble.UserChangeEvent) {
			soundboard.UpdateUsers(gumbleclient)
			soundboard.SaveUsers(datafile)
		},
	})

}

