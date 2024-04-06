package main

type Config struct {
	Endpoint           string   `json:"endpoint" yaml:"endpoint"`
	Username           string   `json:"username" yaml:"username"`
	Password           string   `json:"password" yaml:"password"`
	InscureTLSVerify   bool     `json:"inscure-tls-verify" yaml:"inscure-tls-verify"`
	Debug              bool     `json:"debug" yaml:"debug"`
	Timeout            int      `json:"timeout" yaml:"timeout"`
	Dirs               []Dir    `json:"dirs" yaml:"dirs"`
	Exts               []string `json:"exts" yaml:"exts"`
	CreateSubDirectory bool     `json:"create-sub-directory" yaml:"create-sub-directory"`
	MaxConnections     int      `json:"max-connections" yaml:"max-connections"`
}

type Dir struct {
	LocalDirectory     string   `json:"local-directory" yaml:"local-directory"`
	RemoteDirectories  []string `json:"remote-directories" yaml:"remote-directories"`
	NotRescursive      bool     `json:"not-recursive" yaml:"not-recursive"`
	CreateSubDirectory bool     `json:"create-sub-directory" yaml:"create-sub-directory"`
	Disabled           bool     `json:"disabled" yaml:"disabled"`
	ForceRefresh       bool     `json:"force-refresh" yaml:"force-refresh"`
}
