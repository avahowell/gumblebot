package main

import (
	"bytes"
	"fmt"
	"html/template"
	"github.com/layeh/gumble/gumble_ffmpeg"
	"github.com/layeh/gumble/gumble"
	"github.com/layeh/gumble/gumbleutil"
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

