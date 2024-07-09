package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	sdk "github.com/imshuai/alistsdk-go"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
	"gopkg.in/yaml.v3"
)

const (
	// 定义常量
	NAME        = "AlistAutoStrm"
	DESCRIPTION = "Auto generate .strm file for EMBY or Jellyfin server use Alist API"
	VERSION     = "1.1.0"
)

var (
	logger *StatLogger
	config Config
)

func main() {
	// 隐藏光标
	fmt.Print("\033[?25l")
	// 程序退出时显示光标
	defer func() {
		fmt.Print("\033[?25h")
	}()

	// 初始化一个mpb.Progress实例
	p := mpb.New(mpb.WithAutoRefresh())

	//初始化日志模块
	logger = NewLogger()
	logger.SetFormatter(&Formatter{
		Colored: false,
	})
	logger.SetOutput(p)

	//初始化并设置app实例
	app := cli.NewApp()
	app.Name = NAME
	app.Description = DESCRIPTION
	app.Usage = DESCRIPTION
	app.Version = VERSION
	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:    "config",
			Aliases: []string{"c"},
			Usage:   "Load configuration from `FILE`",
			EnvVars: []string{"ALIST_AUTO_STRM_CONFIG"},
			Value:   "config.json",
		},
	}
	app.Commands = []*cli.Command{
		// fresh-all 命令,重新按Alist服务器数据生成strm文件，不在乎是否已经生成过
		{
			Name:  "fresh-all",
			Usage: "generate all strm files from alist server, whatever the file has been generated or not",
			Action: func(c *cli.Context) error {
				//读取config参数值，并判断传入的是json格式还是yaml格式，再分别使用对应的解析工具解析出Config结构体
				configFile := c.String("config")
				configData, err := os.ReadFile(configFile)
				if err != nil {
					return errors.New("read config file error: " + err.Error())
				}
				config = Config{}
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
				//判断是否使用彩色日志
				if config.ColoredLog {
					logger.SetFormatter(&Formatter{
						Colored: true,
					})
					logger.Info("use colored log")
				}
				//添加进度条
				bar := p.AddBar(0,
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
				// 设置logger的bar
				logger.SetBar(bar)

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
					// 遍历endpoint.Dirs
					for _, dir := range endpoint.Dirs {
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
							m.Run(endpoint.MaxConnections)
							// 增加计数器
							logger.Increment()
						}
					}
				}
				// 进度条完成
				logger.FinishBar()
				logger.Info("generate all strm file done, exit")
				p.Wait()
				return nil
			},
		},
	}
	e := app.Run(os.Args)
	if e != nil {
		logger.Error(e)
	}
}
