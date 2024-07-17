package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	sdk "github.com/imshuai/alistsdk-go"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
	"gopkg.in/yaml.v3"
)

func checkExt(name string, exts []string) bool {
	for _, v := range exts {
		if strings.ToLower(filepath.Ext(name)) == v {
			return true
		}
	}
	return false
}

func urlEncode(s string) string {
	vv := make([]string, 0)
	ss := strings.Split(s, "/")
	for _, v := range ss {
		vv = append(vv, url.PathEscape(v))
	}
	return strings.Join(vv, "/")
}

// 替换字符串中的空格为'-'
func replaceSpaceToDash(s string) string {
	return strings.ReplaceAll(s, " ", "-")
}

func loadConfig(c *cli.Context) error {
	//读取config参数值，并判断传入的是json格式还是yaml格式，再分别使用对应的解析工具解析出Config结构体
	configFile := c.String("config")
	configData, err := os.ReadFile(configFile)
	if err != nil {
		return errors.New("read config file error: " + err.Error())
	}
	config = &Config{}
	//判断传入的是json格式还是yaml格式
	if configFile[len(configFile)-5:] == ".json" {
		//json格式
		err = json.Unmarshal(configData, &config)
		if err != nil {
			return errors.New("unmarshal json type config file error: " + err.Error())
		}
	} else {
		//yaml格式
		err = yaml.Unmarshal(configData, &config)
		if err != nil {
			return errors.New("unmarshal yaml type config file error: " + err.Error())
		}
	}
	return nil
}

func statusBar(p *mpb.Progress) *mpb.Bar {
	return p.AddBar(0,
		//设置进度条前缀
		mpb.PrependDecorators(
			decor.Any(
				func(s decor.Statistics) string {
					return fmt.Sprintf("%s [%7s] Get % 5d files [% 3d/%3d]", NAME, logger.Level.String(), logger.GetCount(), s.Current, s.Total)
				},
				decor.WC{W: 1, C: decor.DSyncWidthR},
			),
		),
		//设置进度条后缀
		mpb.AppendDecorators(
			decor.Elapsed(decor.ET_STYLE_GO, decor.WC{C: decor.DSyncSpace}),
		),
	)
}

func setLogLevel() {
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
}

func fetchRemoteFiles(e Endpoint) ([]*Strm, error) {
	//初始化ALIST Client
	client := sdk.NewClient(e.BaseURL, e.Username, e.Password, e.InscureTLSVerify, config.Timeout)
	u, err := client.Login()
	if err != nil {
		logger.Errorf("login error: %s", err.Error())
	}
	logger.Infof("%s login success, username: %s", e.BaseURL, u.Username)
	return nil, nil
}
