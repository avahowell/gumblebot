package main

import (
	"fmt"
	"regexp"
	"github.com/layeh/gumble/gumble_ffmpeg"
	"github.com/layeh/gumble/gumble"
	"github.com/layeh/gumble/gumbleutil"
	"flag"
	"strings"
)

const datafile = "data"
const UserChangeConnected = 1 << iota
const maxthumbwidth = 200
const image_regex = `[>]https?://.*.(png|jpg|gif|jpeg)([?\s\S]+)?[<]`

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
			imageregex, err := regexp.Compile(image_regex)
			if err != nil {
				panic(err)
			}
			if imageregex.MatchString(e.Message) == true {
				untrimmed := imageregex.FindString(e.Message)
				trimmed := strings.Trim(untrimmed, ">")
				trimmed = strings.Trim(trimmed, "<")
				var thumb MumbleThumbnail
				thumb.MaxWidth = maxthumbwidth
				go thumb.DownloadAndPost(trimmed, gumbleclient)
				return
			}
			if e.Message == "stop" {
				stream.Stop()
				return
			}
			if e.Message == "help" {
				go SendUsage(gumbleclient, soundboard.sounds)
				return
			}
			separated_commands := strings.Split(e.Message, " ")
			if separated_commands[0] == "welcome" {
				if len(separated_commands) == 2 {
					soundboard.SetWelcomeSound(e.Sender.Name, separated_commands[1])
					soundboard.SaveUsers(datafile)
				}
				return
			}
			if e.Message == "sboff" {
				u := soundboard.Users[e.Sender.Name]
				u.SoundboardEnabled = false
				soundboard.Users[e.Sender.Name] = u
				soundboard.SaveUsers(datafile)
				return
			}
			if e.Message == "sbon" {
				u := soundboard.Users[e.Sender.Name]
				u.SoundboardEnabled = true
				soundboard.Users[e.Sender.Name] = u
				soundboard.SaveUsers(datafile)
				return
			}
			soundboard.Play(gumbleclient, stream, e.Message)
		},
		UserChange: func(e *gumble.UserChangeEvent) {
			soundboard.UpdateUsers(gumbleclient)
			soundboard.SaveUsers(datafile)
			if e.Type.Has(UserChangeConnected) == true  {
				soundboard.WelcomeUser(e.User, gumbleclient, stream)
			}
		},
	})

}

