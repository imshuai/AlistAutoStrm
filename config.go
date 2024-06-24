package main

type Config struct {
	Database           string     `json:"database" yaml:"database"`
	Endpoints          []Endpoint `json:"endpoints" yaml:"endpoints"`
	Loglevel           string     `json:"loglevel" yaml:"loglevel"`
	Timeout            int        `json:"timeout" yaml:"timeout"`
	Exts               []string   `json:"exts" yaml:"exts"`
	CreateSubDirectory bool       `json:"create-sub-directory" yaml:"create-sub-directory"`
}

type Dir struct {
	LocalDirectory     string   `json:"local-directory" yaml:"local-directory"`
	RemoteDirectories  []string `json:"remote-directories" yaml:"remote-directories"`
	NotRescursive      bool     `json:"not-recursive" yaml:"not-recursive"`
	CreateSubDirectory bool     `json:"create-sub-directory" yaml:"create-sub-directory"`
	Disabled           bool     `json:"disabled" yaml:"disabled"`
	ForceRefresh       bool     `json:"force-refresh" yaml:"force-refresh"`
}

type Endpoint struct {
	BaseURL string `json:"base-url" yaml:"base-url"`
	//Token            string `json:"token" yaml:"token"`
	Username         string `json:"username" yaml:"username"`
	Password         string `json:"password" yaml:"password"`
	InscureTLSVerify bool   `json:"inscure-tls-verify" yaml:"inscure-tls-verify"`
	Dirs             []Dir  `json:"dirs" yaml:"dirs"`
	MaxConnections   int    `json:"max-connections" yaml:"max-connections"`
}
