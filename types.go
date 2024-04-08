package main

import (
	"os"
)

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
