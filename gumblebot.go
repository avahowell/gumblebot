package main

import (
	"flag"
	"fmt"
	"github.com/layeh/gumble/gumble"
	"github.com/layeh/gumble/gumble_ffmpeg"
	"github.com/layeh/gumble/gumbleutil"
	"strings"
)

const (
	datafile      = "data"
	usersfile     = "users"
	rootuser      = "fighterjet"
	maxthumbwidth = 200
	image_regex   = `[>]https?://.*.(png|jpg|gif|jpeg)([?\s\S]+)?[<]`
)

func main() {
	var sounds_dir = flag.String("sounds", "sounds", "directory where soundboard files are located")
	var volume = flag.Float64("volume", 0.25, "soundboard volume from 0 to 1")
	var stream *gumble_ffmpeg.Stream

	var soundboard Soundboard
	var gumbleclient *gumble.Client
	var admin MumbleAdmin

	soundboard.LoadUsers(datafile)
	soundboard.LoadSounds(*sounds_dir)
	admin.LoadAdminData(usersfile)
	admin.RegisterUser(rootuser, GumblebotAdminRoot)
	var parser MessageParser
	parser.New()

	gumbleutil.Main(func(client *gumble.Client) {
		stream, _ = gumble_ffmpeg.New(client)
		stream.Volume = float32(*volume)
		client.Attach(gumbleutil.AutoBitrate)
		parser.RegisterExpression(image_regex, "Image Thumbnailing",
			func(match string) {
				trimmed := strings.Trim(match, ">")
				trimmed = strings.Trim(trimmed, "<")
				var thumb MumbleThumbnail
				thumb.MaxWidth = maxthumbwidth
				go thumb.DownloadAndPost(trimmed, gumbleclient)
			})
		parser.RegisterCommand("stop", "soundboard stop",
			func(args []string, sender *gumble.User) {
				stream.Stop()
			})
		// TODO parser usage methods, for now just print a static usage template from SendUsage
		parser.RegisterCommand("help", "bot usage command", func(args []string, sender *gumble.User) { go SendUsage(client, soundboard.sounds) })
		parser.RegisterCommand("welcome", "welcome sound",
			func(args []string, sender *gumble.User) {
				soundboard.SetWelcomeSound(sender.Name, args[0])
				soundboard.SaveUsers(datafile)
			})
		parser.RegisterCommand("sbon", "turns soundboard on",
			func(args []string, sender *gumble.User) {
				u := soundboard.Users[sender.Name]
				u.SoundboardEnabled = true
				soundboard.Users[sender.Name] = u
			})
		parser.RegisterCommand("sboff", "turns soundboard off",
			func(args []string, sender *gumble.User) {
				u := soundboard.Users[sender.Name]
				u.SoundboardEnabled = false
				soundboard.Users[sender.Name] = u
			})
		parser.RegisterCommand("whois", "prints admin information about user",
			func(args []string, sender *gumble.User) {
				admin.Whois(sender, args[0], client)
			})
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
			parser.Parse(e.Message, e.Sender)
			soundboard.Play(gumbleclient, stream, e.Message)
		},
		UserChange: func(e *gumble.UserChangeEvent) {
			soundboard.UpdateUsers(gumbleclient)
			soundboard.SaveUsers(datafile)
			if e.Type.Has(gumble.UserChangeConnected) == true {
				soundboard.WelcomeUser(e.User, gumbleclient, stream)
			}
		},
	})

}
