package main

import (
	"encoding/json"
	"errors"
	"os"

	sdk "github.com/imshuai/alistsdk-go"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"
)

var (
	logger *logrus.Logger
	config Config
)

func main() {
	//初始化日志模块
	logger = logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})

	//初始化并设置app实例
	app := cli.NewApp()
	app.Name = "AlistAutoStrm"
	app.Description = "Auto generate .strm file for EMBY or Jellyfin server use Alist API"
	app.Usage = "Auto generate .strm file for EMBY or Jellyfin server use Alist API"
	app.Version = "1.1.0"
	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:    "config",
			Aliases: []string{"c"},
			Usage:   "Load configuration from `FILE`",
			EnvVars: []string{"ALIST_AUTO_STRM_CONFIG"},
			Value:   "config.json",
		},
	}
	app.Action = func(c *cli.Context) error {
		//读取config参数值，并判断传入的是json格式还是yaml格式，再分别使用对应的解析工具解析出Config结构体
		configFile := c.String("config")
		configData, err := os.ReadFile(configFile)
		if err != nil {
			return errors.New("read config file error: " + err.Error())
		}
		config = Config{}
		if configFile[len(configFile)-5:] == ".json" {
			err = json.Unmarshal(configData, &config)
			if err != nil {
				return errors.New("unmarshal json type config file error: " + err.Error())
			}
		} else {
			err = yaml.Unmarshal(configData, &config)
			if err != nil {
				return errors.New("unmarshal yaml type config file error: " + err.Error())
			}
		}
		logger.Infoln("read config file success")
		if config.Debug {
			logger.SetLevel(logrus.DebugLevel)
			logger.Infoln("set log level to debug")
		}

		//输出配置文件调试信息
		logger.Debugf("endpoint: %s", config.Endpoint)
		logger.Debugf("username: %s", config.Username)
		logger.Debugf("password: %s", config.Password)
		logger.Debugf("inscure tls verify: %t", config.InscureTLSVerify)
		logger.Debugf("timeout: %d", config.Timeout)
		logger.Debugf("create sub directory: %t", config.CreateSubDirectory)
		logger.Debugf("exts: %+v", config.Exts)
		logger.Debugf("dirs: %+v", config.Dirs)

		//初始化ALIST Client
		client := sdk.NewClient(config.Endpoint, config.Username, config.Password, config.InscureTLSVerify, config.Timeout)
		u, err := client.Login()
		if err != nil {
			return errors.New("login error: " + err.Error())
		}
		logger.Infoln("login success, username:", u.Username)

		for _, dir := range config.Dirs {
			if dir.Disabled {
				logger.Infof("dir [%s] is disabled, ignore", dir.LocalDirectory)
				continue
			}
			logger.Infoln("start generate .strm file to", dir.LocalDirectory)
			logger.Debugln("create local directory", dir.LocalDirectory)
			err := os.MkdirAll(dir.LocalDirectory, 0666)
			if err != nil {
				return errors.New("create local directory error: " + err.Error())
			}
			for _, rDir := range dir.RemoteDirectories {
				logger.Infof("start get files from %s%s", rDir, func() string {
					if dir.NotRescursive {
						return ""
					} else {
						return " recursively"
					}
				}())
				m := &Mission{
					CurrentRemotePath:    rDir,
					LocalPath:            dir.LocalDirectory,
					Exts:                 config.Exts,
					IsCreateSubDirectory: config.CreateSubDirectory || dir.CreateSubDirectory,
					IsRecursive:          !dir.NotRescursive,
					IsForceRefresh:       dir.ForceRefresh,
					client:               client,
				}
				m.Run(config.MaxConnections)
			}
		}
		logger.Infoln("generate all strm file done, exit")
		return nil
	}
	e := app.Run(os.Args)
	//e := app.Run([]string{"--config", "config.json"})
	if e != nil {
		logger.Error(e)
	}
}
