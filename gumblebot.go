package main

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"github.com/layeh/gumble/gumble_ffmpeg"
	"github.com/layeh/gumble/gumble"
	"github.com/layeh/gumble/gumbleutil"
	"flag"
)


func main() {
	var sounds_dir = flag.String("sounds", "sounds", "directory where soundboard files are located")
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
		value, _ := filepath.Abs(filepath.Join(*sounds_dir, key))
		soundboard[key] = value
	}


	gumbleutil.Main(func(client *gumble.Client) {
		stream, _ = gumble_ffmpeg.New(client)
		stream.Volume = 0.25
		client.Attach(gumbleutil.AutoBitrate)
	}, gumbleutil.Listener{
		Connect: func(e *gumble.ConnectEvent) {
			fmt.Printf("Connected!\n")
		},
		TextMessage: func(e *gumble.TextMessageEvent) {
			if e.Sender == nil {
				return
			}
			if (e.Message == "stop") {
				stream.Stop()
			}
			for key, value := range soundboard {
				if strings.Index(key, e.Message) == 0 {
					stream.Stop()
					stream.Play(value)
				}
			}
		},
	})

}

