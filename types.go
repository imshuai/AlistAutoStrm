package main

import (
	"os"

	sdk "github.com/imshuai/alistsdk-go"
)

type File struct {
	sdk.File
	RemoteDir string `json:"remote_dir"`
	LocalDir  string `json:"local_dir"`
}

type Strm struct {
	Name   string `json:"name"`
	Dir    string `json:"dir"`
	RawURL string `json:"raw_url"`
}

func (s Strm) Save() error {
	err := os.MkdirAll(s.Dir, 0666)
	if err != nil {
		return err
	}
	return os.WriteFile(s.Dir+"/"+s.Name, []byte(s.RawURL), 0666)
}
