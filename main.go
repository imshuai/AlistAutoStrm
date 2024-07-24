package main

import (
	"fmt"
	"os"

	sdk "github.com/imshuai/alistsdk-go"
	"github.com/urfave/cli/v2"
	"github.com/vbauerster/mpb/v8"
)

const (
	// 定义常量
	NAME        = "AlistAutoStrm"
	DESCRIPTION = "Auto generate .strm file for EMBY or Jellyfin server use Alist API"
	VERSION     = "1.1.0"
)

var (
	logger *StatLogger
	config *Config
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
				//从命令行参数读取配置文件信息
				err := loadConfig(c)
				if err != nil {
					return err
				}
				//判断是否使用彩色日志
				if config.ColoredLog {
					logger.SetFormatter(&Formatter{
						Colored: true,
					})
					logger.Info("use colored log")
				}
				//添加进度条
				bar := statusBar(p)
				// 设置logger的bar
				logger.SetBar(bar)

				logger.Info("read config file success")
				logger.Infof("set log level: %s", config.Loglevel)
				// 设置日志等级
				setLogLevel()

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
		{
			Name:  "update",
			Usage: "update strm file with choosed mode",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "mode",
					Usage: "update mode, support: local, remote. when strm content is same but filename changed, local: keep local filename, remote: rename local filename to remote filename",
					Value: "local",
				},
			},
			Action: func(c *cli.Context) error {
				//TODO 实现strm文件更新功能
				//从命令行参数读取配置文件信息
				err := loadConfig(c)
				if err != nil {
					return err
				}
				//判断是否使用彩色日志
				if config.ColoredLog {
					logger.SetFormatter(&Formatter{
						Colored: true,
					})
					logger.Info("use colored log")
				}
				//添加进度条
				bar := statusBar(p)
				// 设置logger的bar
				logger.SetBar(bar)

				logger.Info("read config file success")
				logger.Infof("set log level: %s", config.Loglevel)
				// 设置日志等级
				setLogLevel()

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

				mode := c.String("mode")
				logger.Debugf("update mode: %s", mode)
				localStrms := make(map[string]*Strm, 0)
				remoteStrms := make(map[string]*Strm, 0)
				addStrms := make([]*Strm, 0)
				deleteStrms := make([]*Strm, 0)
				switch mode {
				case "local":
					//TODO 实现本地更新模式
					for _, e := range config.Endpoints {
						for _, v := range fetchLocalFiles(e) {
							localStrms[v.Key()] = v
						}
						for _, v := range fetchRemoteFiles(e) {
							if _, ok := localStrms[v.Key()]; !ok {
								addStrms = append(addStrms, v)
							} else {
								logger.Infof("remote file %s is same with local file %s, ignored.", localStrms[v.Key()].Name, v.Name)
							}
						}
					}
				case "remote":
					//TODO 实现远程更新模式
					for _, e := range config.Endpoints {
						for _, v := range fetchRemoteFiles(e) {
							remoteStrms[v.Key()] = v
						}
						for _, v := range fetchLocalFiles(e) {
							if _, ok := remoteStrms[v.Key()]; !ok {
								deleteStrms = append(deleteStrms, v)
							} else {
								logger.Infof("local file %s is same with remote file %s, ignored.", remoteStrms[v.Key()].Name, v.Name)
							}
						}
					}
				default:
					return fmt.Errorf("invalid update mode: %s", mode)
				}
				for _, v := range addStrms {
					e := v.GenStrm()

					if e != nil {
						logger.Errorf("gen file %s failed: %s", v.Name, e)
					}

					logger.Infof("gen file %s success", v.Dir+"/"+v.Name)
				}

				for _, v := range deleteStrms {
					e := v.Delete()

					if e != nil {
						logger.Errorf("delete file %s failed: %s", v.Name, e)
					}
				}
				logger.Infof("add %d files, delete %d files", len(addStrms), len(deleteStrms))
				return nil
			},
		},
		{
			Name:  "check",
			Usage: "check if strm file is valid",
			Flags: []cli.Flag{
				//添加valid, invalid 两个选项，用于设置对应的数据保存路径
				&cli.StringFlag{
					Name:  "valid",
					Usage: "valid strm list path",
					Value: "valid.csv",
				},
				&cli.StringFlag{
					Name:  "invalid",
					Usage: "invalid strm list path",
					Value: "invalid.csv",
				},
			},
			Action: func(c *cli.Context) error {
				//TODO 实现strm文件有效性校验功能
				return nil
			},
		},
		{
			Name:  "version",
			Usage: "show version",
			Action: func(c *cli.Context) error {
				fmt.Println("version: " + VERSION)
				return nil
			},
		},
	}
	e := app.Run(os.Args)
	if e != nil {
		logger.Error(e)
	}
}
