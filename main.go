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
	logger.SetFormatter(&Formatter{
		Colored: false,
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
		if config.ColoredLog {
			logger.SetFormatter(&Formatter{
				Colored: true,
			})
			logger.Info("use colored log")
		}
		logger.Info("read config file success")
		logger.Infof("set log level: %s", config.Loglevel)
		switch config.Loglevel {
		case "debug", "DEBUG":
			logger.SetLevel(logrus.DebugLevel)
		case "info", "INFO":
			logger.SetLevel(logrus.InfoLevel)
		case "warn", "warning", "WARN":
			logger.SetLevel(logrus.WarnLevel)
		case "error", "ERROR":
			logger.SetLevel(logrus.ErrorLevel)
		case "fatal", "FATAL":
			logger.SetLevel(logrus.FatalLevel)
		case "panic", "PANIC":
			logger.SetLevel(logrus.PanicLevel)
		default:
			logger.SetLevel(logrus.InfoLevel)
		}

		//输出配置文件调试信息
		for _, endpoint := range config.Endpoints {
			logger.Debugf("base url: %s", endpoint.BaseURL)
			//logger.Debugf("token: %s", endpoint.Token)
			logger.Debugf("username: %s", endpoint.Username)
			logger.Debugf("password: %s", endpoint.Password)
			logger.Debugf("inscure tls verify: %t", endpoint.InscureTLSVerify)
			logger.Debugf("dirs: %+v", endpoint.Dirs)
			logger.Debugf("max connections: %d", endpoint.MaxConnections)
		}
		logger.Debugf("timeout: %d", config.Timeout)
		logger.Debugf("create sub directory: %t", config.CreateSubDirectory)
		logger.Debugf("exts: %+v", config.Exts)

		for _, endpoint := range config.Endpoints {
			//开始按配置文件遍历远程目录
			logger.Debugf("start to get files from: %s", endpoint.BaseURL)

			//初始化ALIST Client
			client := sdk.NewClient(endpoint.BaseURL, endpoint.Username, endpoint.Password, endpoint.InscureTLSVerify, config.Timeout)
			u, err := client.Login()
			if err != nil {
				logger.Errorf("login error: %s", err.Error())
				continue
			}
			logger.Infof("%s login success, username: %s", endpoint.BaseURL, u.Username)
			for _, dir := range endpoint.Dirs {
				if dir.Disabled {
					logger.Infof("dir [%s] is disabled", dir.LocalDirectory)
					continue
				}
				logger.Debug("create local directory", dir.LocalDirectory)
				err := os.MkdirAll(dir.LocalDirectory, 0666)
				if err != nil {
					logger.Errorf("create local directory %s error: %s", dir.LocalDirectory, err.Error())
					continue
				}
				for _, remoteDir := range dir.RemoteDirectories {
					m := &Mission{
						CurrentRemotePath:    remoteDir,
						LocalPath:            dir.LocalDirectory,
						Exts:                 config.Exts,
						IsCreateSubDirectory: config.CreateSubDirectory || dir.CreateSubDirectory,
						IsRecursive:          !dir.NotRescursive,
						IsForceRefresh:       dir.ForceRefresh,
						client:               client,
					}
					m.Run(endpoint.MaxConnections)
				}
			}
		}
		logger.Info("generate all strm file done, exit")
		return nil
	}
	e := app.Run(os.Args)
	//e := app.Run([]string{"--config", "config.json"})
	if e != nil {
		logger.Error(e)
	}
}
