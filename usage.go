package main

import (
	"bytes"
	"html/template"
	"github.com/layeh/gumble/gumble"
)

const usageTemplate = `
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
		<li> welcome [sound], gumblebot will welcome you with your sound of choice every time you join the server </li>
    </li>
</ul>`

func SendUsage(client *gumble.Client, soundboard map[string]string) {
	var buffer bytes.Buffer

	outtemplate, err := template.New("help").Parse(usageTemplate)
	if err != nil {
		panic(err)
	}

	err = outtemplate.Execute(&buffer, soundboard )
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
