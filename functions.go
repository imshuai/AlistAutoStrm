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

func fetchRemoteFiles(e Endpoint) []*Strm {
	//初始化ALIST Client
	client := sdk.NewClient(e.BaseURL, e.Username, e.Password, e.InscureTLSVerify, config.Timeout)
	u, err := client.Login()
	if err != nil {
		logger.Errorf("login error: %s", err.Error())
	}
	logger.Infof("%s login success, username: %s", e.BaseURL, u.Username)
	for _, dir := range e.Dirs {
		// 设置总共需要同步的目录数量
		logger.SetTotal(int64(len(dir.RemoteDirectories)) + logger.GetCurrent())
		// 如果目录被禁用，则跳过
		if dir.Disabled {
			logger.Infof("dir [%s] is disabled", dir.LocalDirectory)
			continue
		}
		// 创建本地目录
		logger.Debug("create local directory", dir.LocalDirectory)
		err := os.MkdirAll(dir.LocalDirectory, 0666)
		if err != nil {
			logger.Errorf("create local directory %s error: %s", dir.LocalDirectory, err.Error())
			continue
		}
		// 遍历dir.RemoteDirectories
		for _, remoteDir := range dir.RemoteDirectories {
			// 开始生成strm文件
			logger.Infof("start to generate strm file from remote directory: %s", remoteDir)
			m := &Mission{
				// 当前远程路径
				CurrentRemotePath: remoteDir,
				// 本地路径
				LocalPath: dir.LocalDirectory,
				// 扩展名
				Exts: config.Exts,
				// 是否创建子目录
				IsCreateSubDirectory: config.CreateSubDirectory || dir.CreateSubDirectory,
				// 是否递归
				IsRecursive: !dir.NotRescursive,
				// 是否强制刷新
				IsForceRefresh: dir.ForceRefresh,
				// 客户端
				client: client,
			}
			// 运行
			m.Run(e.MaxConnections)
			// 增加计数器
			logger.Increment()
		}
	}
	return nil
}

func fetchLocalFiles(e Endpoint) []*Strm {
	// TODO 读取本地已有strm文件
	strms := make([]*Strm, 0)
	for _, dir := range e.Dirs {
		// 遍历路径下所有strm文件，包括子目录中
		files, err := filepath.Glob(filepath.Join(dir.LocalDirectory, "*.strm"))

		if err != nil {
			// 读取本地目录出错，记录错误日志
			logger.Errorf("read local directory %s error: %s", dir.LocalDirectory, err.Error())
			continue
		}
		for _, file := range files {
			// 读取strm文件，返回Strm结构体
			strm := readStrmFile(file)
			// 将读取的strm文件添加到strms切片中
			strms = append(strms, strm)
		}
	}
	return strms
}

func readStrmFile(file string) *Strm {
	// TODO 读取strm文件
	strm := &Strm{}
	strm.Name = filepath.Base(file)
	strm.Dir = filepath.Dir(file)
	strm.RawURL = func() string {
		byts, err := os.ReadFile(file)
		if err != nil {
			return ""
		}
		return string(byts)
	}()
	return strm
}
