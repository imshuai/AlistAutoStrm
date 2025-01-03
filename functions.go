package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	sdk "github.com/imshuai/alistsdk-go"
	"github.com/sirupsen/logrus"
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

func urlDecode(s string) string {
	vv := make([]string, 0)
	ss := strings.Split(s, "/")
	for _, v := range ss {
		vvv, _ := url.PathUnescape(v)
		vv = append(vv, vvv)
	}
	return strings.Join(vv, "/")
}

func loadConfig(configFile string) (*Config, error) {
	//读取config参数值，并判断传入的是json格式还是yaml格式，再分别使用对应的解析工具解析出Config结构体
	configData, err := os.ReadFile(configFile)
	if err != nil {
		return nil, errors.New("read config file error: " + err.Error())
	}
	config := &Config{}
	//判断传入的是json格式还是yaml格式
	if configFile[len(configFile)-5:] == ".json" {
		//json格式
		err = json.Unmarshal(configData, &config)
		if err != nil {
			return nil, errors.New("unmarshal json type config file error: " + err.Error())
		}
	} else {
		//yaml格式
		err = yaml.Unmarshal(configData, &config)
		if err != nil {
			return nil, errors.New("unmarshal yaml type config file error: " + err.Error())
		}
	}
	return config, nil
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
	switch strings.ToLower(config.Loglevel) {
	case "trace":
		logger.SetLevel(logrus.TraceLevel)
	case "debug":
		logger.SetLevel(logrus.DebugLevel)
	case "info":
		logger.SetLevel(logrus.InfoLevel)
	case "warn", "warning":
		logger.SetLevel(logrus.WarnLevel)
	case "error":
		logger.SetLevel(logrus.ErrorLevel)
	case "fatal":
		logger.SetLevel(logrus.FatalLevel)
	case "panic":
		logger.SetLevel(logrus.PanicLevel)
	default:
		logger.SetLevel(logrus.InfoLevel)
	}
}

func fetchRemoteFiles(e Endpoint) []*Strm {
	//初始化ALIST Client
	var client *sdk.Client
	if e.Token != "" {
		client = sdk.NewClientWithToken(e.BaseURL, e.Token, e.InscureTLSVerify, config.Timeout)
	} else {
		client = sdk.NewClient(e.BaseURL, e.Username, e.Password, e.InscureTLSVerify, config.Timeout)
	}
	//登录
	u, err := client.Login()
	if err != nil {
		logger.Errorf("[MAIN]: login error: %s", err.Error())
		return nil
	}
	logger.Infof("[MAIN]: %s login success, username: %s", e.BaseURL, u.Username)
	strms := make([]*Strm, 0)
	for _, dir := range e.Dirs {
		// 设置总共需要同步的目录数量
		logger.SetTotal(int64(len(dir.RemoteDirectories)) + logger.GetCurrent())
		// 如果目录被禁用，则跳过
		if dir.Disabled {
			logger.Infof("[MAIN]: dir [%s] is disabled", dir.LocalDirectory)
			continue
		}
		// 遍历dir.RemoteDirectories
		for _, remoteDir := range dir.RemoteDirectories {
			// 开始生成strm文件
			logger.Infof("[MAIN]: fetch strm info from remote directory: %s", remoteDir)
			m := &Mission{
				// 服务器地址
				BaseURL: e.BaseURL,
				// 当前远程路径
				CurrentRemotePath: remoteDir,
				// 本地路径
				LocalPath: dir.LocalDirectory,
				// 扩展名
				Exts: config.Exts,
				// 额外扩展名
				AltExts: config.AltExts,
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
			strms = append(strms, m.GetAllStrm(e.MaxConnections)...)
			// 增加计数器
			logger.Increment()
		}
	}
	return strms
}

func fetchLocalFiles(e Endpoint) []*Strm {
	// TODO 读取本地已有strm文件
	strms := make([]*Strm, 0)
	for _, dir := range e.Dirs {
		if dir.Disabled {
			logger.Infof("[MAIN]: dir [%s] is disabled", dir.LocalDirectory)
			continue
		}
		logger.Infof("[MAIN]: reading local directory %s", dir.LocalDirectory)
		// 遍历路径下所有strm文件，包括子目录中
		files := make([]string, 0)
		err := filepath.Walk(dir.LocalDirectory, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && filepath.Ext(path) == ".strm" {
				files = append(files, path)
			}
			return nil
		})
		if err != nil {
			// 读取本地目录出错，记录错误日志
			logger.Warnf("[MAIN]: read local directory %s error: %s", dir.LocalDirectory, err.Error())
			continue
		}
		logger.Infof("[MAIN]: find %d strm files", len(files))
		for _, file := range files {
			// 读取strm文件，返回Strm结构体
			strm := readStrmFile(file)
			logger.Tracef("[MAIN]: read local strm file %s, url: %s", file, strm.RawURL)
			// 将读取的strm文件添加到strms切片中
			strms = append(strms, strm)
		}
		time.Sleep(time.Millisecond * 200)
	}
	return strms
}

func readStrmFile(file string) *Strm {
	// TODO 读取strm文件
	strm := &Strm{}
	strm.Name = filepath.Base(file)
	strm.LocalDir = filepath.Dir(file)
	strm.RawURL = func() string {
		byts, err := os.ReadFile(file)
		if err != nil {
			return ""
		}
		//返回的字符串应该有且只有一行，且不会以\n或者\r\n结束
		return strings.TrimRight(strings.Split(string(byts), "\n")[0], "\r")
	}()
	strm.RemoteDir = func() string {
		logger.Tracef("[MAIN]: parse remote directory from strm: %s url: %s", file, strm.RawURL)
		paths := strings.Split(strings.Split(strm.RawURL, "/d/")[1], "/")
		paths = paths[:len(paths)-1]
		str := urlDecode("/" + strings.Join(paths, "/"))
		logger.Debugf("[MAIN]: remote directory: %s", str)
		return str
	}()
	return strm
}

func PrintDebugInfo() {
	//输出配置文件调试信息
	for _, endpoint := range config.Endpoints {
		logger.Debugf("[MAIN]: base url: %s", endpoint.BaseURL)
		logger.Debugf("[MAIN]: token: %s", endpoint.Token)
		logger.Debugf("[MAIN]: username: %s", endpoint.Username)
		logger.Debugf("[MAIN]: password: %s", endpoint.Password)
		logger.Debugf("[MAIN]: inscure tls verify: %t", endpoint.InscureTLSVerify)
		logger.Debugf("[MAIN]: dirs: %+v", endpoint.Dirs)
		logger.Debugf("[MAIN]: max connections: %d", endpoint.MaxConnections)
	}
	logger.Debugf("[MAIN]: timeout: %d", config.Timeout)
	logger.Debugf("[MAIN]: create sub directory: %t", config.CreateSubDirectory)
	logger.Debugf("[MAIN]: exts: %+v", config.Exts)
}
