package main

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"github.com/layeh/gumble/gumble_ffmpeg"
	"github.com/layeh/gumble/gumble"
	"github.com/layeh/gumble/gumbleutil"
)


func main() {
	soundboard := make(map[string]string)

	// Populate soundboard with files in $pwd/sounds/
	files, _ := ioutil.ReadDir("./sounds")
	for _, f := range files {
		key := f.Name()
		value, _ := filepath.Abs(filepath.Join("sounds/", key))
		soundboard[key] = value
	}

	var stream *gumble_ffmpeg.Stream

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

