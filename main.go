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
	VERSION     = "1.2.0"
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
		if db != nil {
			db.Close()
		}
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

				logger.Info("[MAIN]: read config file success")
				logger.Infof("[MAIN]: set log level: %s", config.Loglevel)
				// 设置日志等级
				setLogLevel()

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

				for _, endpoint := range config.Endpoints {
					//开始按配置文件遍历远程目录
					logger.Debugf("[MAIN]: start to get files from: %s", endpoint.BaseURL)

					//初始化ALIST Client
					var client *sdk.Client
					if endpoint.Token != "" {
						client = sdk.NewClientWithToken(endpoint.BaseURL, endpoint.Token, endpoint.InscureTLSVerify, config.Timeout)
					} else {
						client = sdk.NewClient(endpoint.BaseURL, endpoint.Username, endpoint.Password, endpoint.InscureTLSVerify, config.Timeout)
					}
					// 登录
					u, err := client.Login()
					if err != nil {
						logger.Errorf("[MAIN]: login error: %s", err.Error())
						continue
					}
					logger.Infof("[MAIN]: %s login success, username: %s", endpoint.BaseURL, u.Username)
					// 遍历endpoint.Dirs
					for _, dir := range endpoint.Dirs {
						// 设置总共需要同步的目录数量
						logger.SetTotal(int64(len(dir.RemoteDirectories)) + logger.GetCurrent())
						// 如果目录被禁用，则跳过
						if dir.Disabled {
							logger.Infof("[MAIN]: dir [%s] is disabled", dir.LocalDirectory)
							continue
						}
						// 创建本地目录
						logger.Debug("[MAIN]: create local directory", dir.LocalDirectory)
						err := os.MkdirAll(dir.LocalDirectory, 0666)
						if err != nil {
							logger.Errorf("[MAIN]: create local directory %s error: %s", dir.LocalDirectory, err.Error())
							continue
						}
						// 遍历dir.RemoteDirectories
						for _, remoteDir := range dir.RemoteDirectories {
							// 开始生成strm文件
							logger.Infof("[MAIN]: start to generate strm file from remote directory: %s", remoteDir)
							m := &Mission{
								// 服务器地址
								BaseURL: endpoint.BaseURL,
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
				logger.Info("[MAIN]: generate all strm file done, exit")
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
				&cli.BoolFlag{
					Name:  "no-incremental-update",
					Usage: "when this flag is set, will not use incremental update, will update all files",
					Value: false,
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
					logger.Info("[MAIN]: use colored log")
				}
				//添加进度条
				bar := statusBar(p)
				// 设置logger的bar
				logger.SetBar(bar)

				logger.Info("[MAIN]: read config file success")
				logger.Infof("[MAIN]: set log level: %s", config.Loglevel)
				// 设置日志等级
				setLogLevel()

				//输出配置文件调试信息
				for _, endpoint := range config.Endpoints {
					logger.Debugf("[MAIN]: base url: %s", endpoint.BaseURL)
					logger.Debugf("token: %s", endpoint.Token)
					logger.Debugf("[MAIN]: username: %s", endpoint.Username)
					logger.Debugf("[MAIN]: password: %s", endpoint.Password)
					logger.Debugf("[MAIN]: inscure tls verify: %t", endpoint.InscureTLSVerify)
					logger.Debugf("[MAIN]: dirs: %+v", endpoint.Dirs)
					logger.Debugf("[MAIN]: max connections: %d", endpoint.MaxConnections)
				}
				logger.Debugf("[MAIN]: timeout: %d", config.Timeout)
				logger.Debugf("[MAIN]: create sub directory: %t", config.CreateSubDirectory)
				logger.Debugf("[MAIN]: exts: %+v", config.Exts)

				mode := c.String("mode")
				logger.Debugf("[MAIN]: update mode: %s", mode)
				config.isIncrementalUpdate = !c.Bool("no-incremental-update") //是否使用增量更新
				logger.Debugf("[MAIN]: incremental update: %t", config.isIncrementalUpdate)
				localStrms := make(map[string]*Strm, 0)
				remoteStrms := make(map[string]*Strm, 0)
				addStrms := make([]*Strm, 0)
				deleteStrms := make([]*Strm, 0)
				ignored, added, deleted := 0, 0, 0
				switch mode {
				case "local":
					//TODO 实现本地更新模式
					for _, e := range config.Endpoints {
						localData := fetchLocalFiles(e)
						logger.Infof("[MAIN]: fetched %d local files", len(localData))
						for _, v := range localData {
							localStrms[v.Key()] = v
						}
						remoteData := fetchRemoteFiles(e)
						logger.Infof("[MAIN]: fetched %d remote files", len(remoteData))
						for _, v := range remoteData {
							if _, ok := localStrms[v.Key()]; !ok {
								addStrms = append(addStrms, v)
								logger.Debugf("[MAIN]: %s 已加入待保存列表", v.Name)
								logger.Tracef("[MAIN]: raw_url: %s", v.RawURL)
							} else {
								ignored++
								logger.Debugf("[MAIN]: %s already exits, ignored.", v.Name)
								logger.Tracef("[MAIN]: local content: %s", localStrms[v.Key()].RawURL)
								logger.Tracef("[MAIN]: remote content: %s", v.RawURL)
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
								ignored++
								logger.Infof("[MAIN]: %s already exits, ignored.", v.Name)
								logger.Tracef("[MAIN]: local content: %s", localStrms[v.Key()].RawURL)
								logger.Tracef("[MAIN]: remote content: %s", v.RawURL)
							}
						}
					}
				default:
					return fmt.Errorf("[MAIN]: invalid update mode: %s", mode)
				}
				for _, v := range addStrms {
					var e error
					if mode == "local" {
						e = v.GenStrm(false)
					} else {
						e = v.GenStrm(true)
					}

					if e != nil {
						logger.Warnf("[MAIN]: generate file %s failed: %s", v.Name, e)
						continue
					}
					added++
					logger.Infof("[MAIN]: generate file %s success", v.LocalDir+"/"+v.Name)
				}

				for _, v := range deleteStrms {
					e := v.Delete()

					if e != nil {
						logger.Warnf("[MAIN]: delete file %s failed: %s", v.Name, e)
						continue
					}
					deleted++
				}
				logger.Infof("[MAIN]: want to add %d files, want to delete %d files", len(addStrms), len(deleteStrms))
				logger.Infof("[MAIN]: ignored %d files, added %d files, deleted %d files", ignored, added, deleted)
				logger.FinishBar()
				p.Wait()
				return nil
			},
		},
		{
			Name:  "update-database",
			Usage: "clean database and get all local strm files stored in database",
			Action: func(c *cli.Context) error {
				err := loadConfig(c)
				if err != nil {
					return err
				}
				records := make(map[string]int, 0)
				for _, e := range config.Endpoints {
					strms := fetchLocalFiles(e)
					for _, v := range strms {
						records[v.RemoteDir] = 0
					}
				}
				err = SaveRecordCollection(records)
				if err != nil {
					return err
				}
				logger.Infof("[MAIN]: database has been cleaned, and %d records saved", len(records))
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
