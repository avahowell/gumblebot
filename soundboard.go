package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/layeh/gumble/gumble"
	"github.com/layeh/gumble/gumble_ffmpeg"
	"io/ioutil"
	"path/filepath"
	"strings"
	"time"
)

type SoundboardUser struct {
	SoundboardEnabled bool
	SoundboardVolume  float64
	WelcomeSound      string
}
type Soundboard struct {
	Users  map[string]SoundboardUser
	sounds map[string]string
}

func (s *Soundboard) LoadSounds(path string) {
	s.sounds = make(map[string]string)
	files, err := ioutil.ReadDir(path)
	if err != nil {
		fmt.Println(err)
		return
	}
	for _, f := range files {
		key := f.Name()
		if filepath.Ext(key) == ".mp3" || filepath.Ext(key) == ".ogg" || filepath.Ext(key) == ".wav" {
			value, _ := filepath.Abs(filepath.Join(path, key))
			s.sounds[key] = value
		}
	}
}
func (s *Soundboard) LoadUsers(datapath string) {
	iobuffer, err := ioutil.ReadFile(datapath)
	if err != nil {
		fmt.Println(err)
		return
	}
	buffer := bytes.NewBuffer(iobuffer)
	dec := gob.NewDecoder(buffer)
	err = dec.Decode(&s.Users)
	if err != nil {
		panic(err)
	}
}
func (s *Soundboard) SaveUsers(datapath string) {
	var buffer bytes.Buffer
	enc := gob.NewEncoder(&buffer)
	err := enc.Encode(s.Users)
	if err != nil {
		panic(err)
	}
	err = ioutil.WriteFile(datapath, buffer.Bytes(), 0755)
	if err != nil {
		panic(err)
	}
}
func (s *Soundboard) UpdateUsers(client *gumble.Client) {
	for _, user := range client.Users {
		if _, ok := s.Users[user.Name]; ok {
			continue
		}
		s.Users[user.Name] = SoundboardUser{false, 0.5, ""}
	}
}
func (s *Soundboard) WelcomeUser(user *gumble.User, client *gumble.Client, stream *gumble_ffmpeg.Stream) {
	if _, ok := s.Users[user.Name]; ok {
		if len(s.Users[user.Name].WelcomeSound) > 0 {
			vtarget := &gumble.VoiceTarget{}
			vtarget.ID = 1
			vtarget.AddUser(user)
			for stream.IsPlaying() {
				time.Sleep(50 * time.Millisecond)
			}
			client.Send(vtarget)
			client.VoiceTarget = vtarget
			stream.Play(s.Users[user.Name].WelcomeSound)
		}
	}
}
func (s *Soundboard) SetWelcomeSound(username string, sound string) {
	for key, value := range s.sounds {
		if strings.Index(key, sound) == 0 {
			if _, ok := s.Users[username]; ok {
				u := s.Users[username]
				u.WelcomeSound = value
				s.Users[username] = u
				return
			}
		}
	}
}
func (s *Soundboard) Play(client *gumble.Client, stream *gumble_ffmpeg.Stream, sound string) {
	for soundname, soundpath := range s.sounds {
		if strings.Index(soundname, sound) == 0 {
			vtarget := &gumble.VoiceTarget{}
			vtarget.ID = 1
			for username, sb := range s.Users {
				userstruct := client.Users.Find(username)
				if userstruct != nil && sb.SoundboardEnabled == true {
					vtarget.AddUser(userstruct)
				}
			}
			stream.Stop()
			client.Send(vtarget)
			client.VoiceTarget = vtarget
			stream.Play(soundpath)
		}
	}
}
