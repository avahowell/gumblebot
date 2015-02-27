package main

import ( "github.com/nfnt/resize"
		 "image"
		 "image/jpeg"
		 _"image/png"
		 _"image/gif"
		 "bytes"
		 "net/http"
		 "encoding/base64"
		 "html/template"
		 "github.com/layeh/gumble/gumble"
		 "fmt" )
const mumbleImageTemplate = `
<a href="{{.Link}}"><img src="data:image/jpeg;base64,{{.Data}}"/></a>`

func orPanic(err error) {
	if err != nil {
		panic(err)
	}
}
type ThumbnailContext struct {
	Data string
	Link string
}

type MumbleThumbnail struct {
	Base64Data string
	MaxWidth uint
	Source string
}

func (m *MumbleThumbnail) Download(url string) {
	res, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
		return
	}
	m.Source = url

	image, _, err := image.Decode(res.Body)
	if err != nil {
		fmt.Println(err)
		return
	}

	buf := new(bytes.Buffer)
	resized := resize.Resize(m.MaxWidth, 0, image, resize.Lanczos3)
	err = jpeg.Encode(buf, resized, nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	m.Base64Data = base64.StdEncoding.EncodeToString(buf.Bytes())
}
func (m *MumbleThumbnail) Post (client *gumble.Client) {
	var buffer bytes.Buffer

	outTemplate, err := template.New("img").Parse(mumbleImageTemplate)
	orPanic(err)

	fmt.Println(m.Source)
	err = outTemplate.Execute(&buffer, ThumbnailContext{m.Base64Data, m.Source} )
	orPanic(err)

	message := gumble.TextMessage{
		Channels: []*gumble.Channel {
			client.Self.Channel,
		},
		Message: buffer.String(),
	}
	client.Send(&message)
}
func (m *MumbleThumbnail) DownloadAndPost(url string, client *gumble.Client) {
	m.Download(url)
	m.Post(client)
}
