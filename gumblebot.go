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
	maxthumbwidth = 200
	image_regex   = `[>]https?://.*.(png|jpg|gif|jpeg)([?\s\S]+)?[<]`
)

func main() {
	var sounds_dir = flag.String("sounds", "sounds", "directory where soundboard files are located")
	var volume = flag.Float64("volume", 0.25, "soundboard volume from 0 to 1")
	var rootuser = flag.String("root", "fighterjet", "the root user for the gumblebot admin subsystem")
	var stream *gumble_ffmpeg.Stream

	var soundboard Soundboard
	var gumbleclient *gumble.Client
	var admin MumbleAdmin

	soundboard.LoadUsers(datafile)
	soundboard.LoadSounds(*sounds_dir)
	admin.LoadAdminData(usersfile)
	admin.RegisterUser(*rootuser, GumblebotRoot)
	var parser MessageParser
	parser.New()

	gumbleutil.Main(func(client *gumble.Client) {
		stream, _ = gumble_ffmpeg.New(client)
		stream.Volume = float32(*volume)
		client.Attach(gumbleutil.AutoBitrate)
		admin.Attach(client)
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
		parser.RegisterCommand("welcome", "welcome sound",
			func(args []string, sender *gumble.User) {
				if len(args) < 1 {
					SendMumbleMessage(parser.Commands["welcome"].Usage, client, client.Self.Channel)
					return
				}
				soundboard.SetWelcomeSound(sender.Name, args[0])
				soundboard.SaveUsers(datafile)
			})
		parser.RegisterCommand("sbon", "opt-in to soundboard",
			func(args []string, sender *gumble.User) {
				u := soundboard.Users[sender.Name]
				u.SoundboardEnabled = true
				soundboard.Users[sender.Name] = u
			})
		parser.RegisterCommand("sboff", "opt-out to soundboard",
			func(args []string, sender *gumble.User) {
				u := soundboard.Users[sender.Name]
				u.SoundboardEnabled = false
				soundboard.Users[sender.Name] = u
			})
		parser.RegisterCommand("whois", "prints admin information about user",
			func(args []string, sender *gumble.User) {
				admin.Whois(sender, args[0])
			})
		parser.RegisterCommand("poke", "pokes a user",
			func(args []string, sender *gumble.User) {
				if len(args) < 1 {
					SendMumbleMessage(parser.Commands["poke"].Usage, client, client.Self.Channel)
					return
				}
				admin.Poke(sender, args[0])
			})
		parser.RegisterCommand("sbusers", "prints soundboard users",
			func(args []string, sender *gumble.User) {
				// TODO
			})
		parser.RegisterCommand("move", "move [users...] [channel]",
			func(args []string, sender *gumble.User) {
				if len(args) < 2 {
					SendMumbleMessage(parser.Commands["move"].Usage, client, client.Self.Channel)
					return
				}
				admin.Move(sender, args[len(args)-1], args[:len(args)-1])
			})
		parser.RegisterCommand("kick", "kick [user] [reason]",
			func(args []string, sender *gumble.User) {
				if len(args) < 2 {
					SendMumbleMessage(parser.Commands["kick"].Usage, client, client.Self.Channel)
					return
				}
				admin.Kick(sender, args[0], args[1])
			})
		parser.RegisterCommand("ban", "ban [user] [reason]",
			func(args []string, sender *gumble.User) {
				if len(args) <2 {
					SendMumbleMessage(parser.Commands["ban"].Usage, client, client.Self.Channel)
					return
				}
				admin.Ban(sender, args[0], args[1])
			})
		parser.RegisterCommand("register", "register [user] [(user, moderator, root)] registers a user as one of the following user, moderator, root",
			func(args []string, sender *gumble.User) {
				if sender_admin, ok := admin.Users[sender.Name]; ok {
					if sender_admin.RegisterAllowed != true {
						SendMumbleMessage(permissiondenied, client, client.Self.Channel)
						return
					}
					if len(args) < 2 {
						SendMumbleMessage(parser.Commands["register"].Usage, client, client.Self.Channel)
						return
					}
					var level string
					switch args[1] {
					case GumblebotRoot:
						level = GumblebotRoot
					case GumblebotModerator:
						level = GumblebotModerator
					case GumblebotUser:
						level = GumblebotUser
					default:
						SendMumbleMessage(parser.Commands["register"].Usage, client, client.Self.Channel)
						return
					}
					targetuser := admin.search_mumble_users_substring(args[0])
					if targetuser != nil {
						admin.RegisterUser(targetuser.Name, level)
						admin.SaveAdminData(usersfile)
					}
				}
			})
		parser.RegisterCommand("help", "prints usage",
			func(args []string, sender *gumble.User) {
				textmessage := gumble.TextMessage{
					Channels: []*gumble.Channel{
						sender.Channel,
					},
					Message: parser.Usage(),
				}
				client.Send(&textmessage)
			})
		parser.RegisterCommand("play", "plays a sound from the soundboard",
			func(args []string, sender *gumble.User) {
				if len(args) < 1 {
					return
				}
				go soundboard.Play(client, stream, args[0])
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
		},
		PermissionDenied: func(e *gumble.PermissionDeniedEvent) {
			fmt.Println(e)
		},
		UserChange: func(e *gumble.UserChangeEvent) {
			soundboard.UpdateUsers(gumbleclient)
			soundboard.SaveUsers(datafile)
			if e.Type.Has(gumble.UserChangeConnected) == true {
				go soundboard.WelcomeUser(e.User, gumbleclient, stream)
			}
		},
	})

}
