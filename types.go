package main

import (
	"crypto/md5"
	"fmt"
	"os"
)

type Strm struct {
	Name   string `json:"name"`
	Dir    string `json:"dir"`
	RawURL string `json:"raw_url"`
}

func (s Strm) Key() string {
	byts := md5.Sum([]byte(s.RawURL))
	return fmt.Sprintf("%x", byts)
}

func (s Strm) Save() error {
	return nil
}

func (s Strm) GenStrm() error {
	err := os.MkdirAll(s.Dir, 0666)
	if err != nil {
		return err
	}
	return os.WriteFile(s.Dir+"/"+s.Name, []byte(s.RawURL), 0666)
}
