# gumblebot
A mumble bot using gumble

Features
------------
- Administration system (kick, ban, move)
- Automatic image parsing and thumbnailing
- Opt-in soundboard system, with welcome sounds
- Fuzzy (strings.Index == 0) chat command parser

Installation
-----------
Requirements:
- ffmpeg
- libopus
- Go, gumble

0. Install dependencies, set $GOPATH and $GOROOT

1. Clone into gumblebot/
```
git clone https://github.com/johnathanhowell/gumblebot
```
2. Use go to install go deps
```
cd gumblebot/
go get ./...
```
Usage
-----------
Connect to locahost:1337 as gumblebot, with the root administrator defined as "luser", and "sandstorm.mp3" in the "sounds" folder
```
./gumblebot -root="luser" -server=localhost:1337 -sounds="sounds" -username="gumblebot"
```

(From mumble chat, in same channel as bot) opt-in to soundboard 
```
sbon
```
Play sandstorm! (dudududu)
```
play sand
```
See "help" for more chat commands.
```
help
```

Register custom chat commands:
- gumblebot.go:
```
parser.RegisterCommand("hello", "hello world",
	func(args []string, sender *gumble.User) {
		SendMumbleMessage(fmt.Sprintf("Hello, %s", sender.Name), client, client.Self.Channel)
})
```
Have fun!

License
-----------
The MIT License (MIT)
