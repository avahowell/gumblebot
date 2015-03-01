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
	admin.RegisterUser(rootuser, GumblebotRoot)
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
				admin.Whois(sender, args[0], client)
			})
		parser.RegisterCommand("poke", "pokes a user",
			func(args []string, sender *gumble.User) {
				if len(args) < 1 {
					SendMumbleMessage(parser.Commands["poke"].Usage, client, client.Self.Channel)
					return
				}
				pokestring := fmt.Sprintf("%s poked you!", sender.Name)
				targetuser := search_mumble_users_substring(args[0], client)
				if targetuser != nil {
					SendMumbleMessageTo(targetuser, pokestring, client)
				} else {
					notfound := fmt.Sprintf("%s not found!", args[0])
					SendMumbleMessage(notfound, client, client.Self.Channel)
				}
			})
		parser.RegisterCommand("sbusers", "prints soundboard users",
			func(args []string, sender *gumble.User) {
				// TODO
			})
		parser.RegisterCommand("register", "register [user] [(user, moderator, root)] registers a user as one of the followingL user, moderator, root",
			func(args []string, sender *gumble.User) {
				if sender_admin, ok := admin.Users[sender.Name]; ok {
					if sender_admin.AccessLevel >= GumblebotRoot {
						fmt.Println(args)
						if len(args) < 2 {
							SendMumbleMessage(parser.Commands["register"].Usage, client, client.Self.Channel)
							return
						}
						targetuser := search_mumble_users_substring(args[0], client)
						if targetuser != nil {
							switch args[1] {
							case "user":
								admin.Users[targetuser.Name] = AdminUser{targetuser.Name, GumblebotUser}
							case "moderator":
								admin.Users[targetuser.Name] = AdminUser{targetuser.Name, GumblebotModerator}
							case "root":
								admin.Users[targetuser.Name] = AdminUser{targetuser.Name, GumblebotRoot}
							default:
								SendMumbleMessage(parser.Commands["register"].Usage, client, client.Self.Channel)
							}
							admin.SaveAdminData(usersfile)
						}
					} else {
						SendMumbleMessage(permissiondenied, client, client.Self.Channel)
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
		UserChange: func(e *gumble.UserChangeEvent) {
			soundboard.UpdateUsers(gumbleclient)
			soundboard.SaveUsers(datafile)
			if e.Type.Has(gumble.UserChangeConnected) == true {
				go soundboard.WelcomeUser(e.User, gumbleclient, stream)
			}
		},
	})

}
