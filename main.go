package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/boltdb/bolt"
	"github.com/urfave/cli/v2"
	"github.com/vbauerster/mpb/v8"
)

const (
	NAME        = "AlistAutoStrm"
	DESCRIPTION = "Auto generate .strm file for EMBY or Jellyfin server use Alist API"
	VERSION     = "1.2.5"
)

var (
	logger *StatLogger
	config *Config
	db     *bolt.DB
)

func main() {
	fmt.Print("\033[?25l")
	defer func() {
		fmt.Print("\033[?25h")
		if db != nil {
			db.Close()
		}
		time.Sleep(time.Second)
	}()

	p := mpb.New(mpb.WithAutoRefresh())

	logger = NewLogger()
	logger.SetFormatter(&Formatter{
		Colored: false,
	})
	logger.SetOutput(p)

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
			Action: func(c *cli.Context, path string) error {
				if c.Command.Name == "version" {
					return nil
				}
				var err error
				config, err = loadConfig(path)
				if err != nil {
					logger.Errorf("[MAIN]: load config error: %s", err.Error())
					return err
				}
				if config.ColoredLog {
					logger.SetFormatter(&Formatter{
						Colored: true,
					})
					logger.Info("[MAIN]: use colored log")
				}
				db, err = bolt.Open(config.Database, 0600, nil)
				if err != nil {
					logger.Errorf("[MAIN]: open database error: %s", err.Error())
					return err
				}
				if config.LogFile != "" {
					f, err := os.OpenFile(config.LogFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
					if err != nil {
						logger.Errorf("[MAIN]: open log file error: %s", err.Error())
						return err
					}
					mwriter := io.MultiWriter(os.Stdout, f)
					logger.SetOutput(mwriter)
				}
				logger.Info("[MAIN]: read config file success")
				logger.Infof("[MAIN]: set log level: %s", config.Loglevel)
				setLogLevel()
				return nil
			},
		},
	}
	app.Commands = []*cli.Command{
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
				var err error
				bar := statusBar(p)
				logger.SetBar(bar)

				PrintDebugInfo()

				mode := c.String("mode")
				logger.Debugf("[MAIN]: update mode: %s", mode)
				config.isIncrementalUpdate = !c.Bool("no-incremental-update")
				logger.Debugf("[MAIN]: incremental update: %t", config.isIncrementalUpdate)
				config.records, err = GetRecordCollection()
				if err != nil {
					err := errors.New("get record collection error: " + err.Error())
					logger.Errorf("[MAIN]: %s", err.Error())
					return err
				}
				localStrms := make(map[string]*Strm, 0)
				remoteStrms := make(map[string]*Strm, 0)
				addStrms := make([]*Strm, 0)
				deleteStrms := make([]*Strm, 0)
				ignored, added, deleted := 0, 0, 0
				switch mode {
				case "local":
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
					err := fmt.Errorf("invalid update mode: %s", mode)
					logger.Errorf("[MAIN]: %s", err.Error())
					return err
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
					config.records[v.RemoteDir] = 0
					added++
					logger.Infof("[MAIN]: generate file %s success", v.LocalDir+"/"+v.Name)
				}

				for _, v := range deleteStrms {
					e := v.Delete()

					if e != nil {
						logger.Warnf("[MAIN]: delete file %s failed: %s", v.Name, e)
						continue
					}
					delete(config.records, v.RemoteDir)
					deleted++
				}
				logger.Infof("[MAIN]: want to add %d files, want to delete %d files", len(addStrms), len(deleteStrms))
				logger.Infof("[MAIN]: ignored %d files, added %d files, deleted %d files", ignored, added, deleted)
				if err := SaveRecordCollection(config.records); err != nil {
					logger.Errorf("[MAIN]: save record collection failed: %s", err)
					return err
				}
				logger.FinishBar()
				p.Wait()
				return nil
			},
		},
		{
			Name:  "update-database",
			Usage: "clean database and get all local strm files stored in database",
			Action: func(c *cli.Context) error {
				PrintDebugInfo()

				records := make(map[string]int, 0)
				for _, e := range config.Endpoints {
					strms := fetchLocalFiles(e)
					for _, v := range strms {
						records[v.RemoteDir] = 0
					}
				}
				logger.Infof("[MAIN]: %d records found", len(records))
				logger.Tracef("[MAIN]: records: %+v", records)
				if err := SaveRecordCollection(records); err != nil {
					logger.Errorf("[MAIN]: save record collection failed: %s", err)
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
				PrintDebugInfo()

				localstrms := func() []*Strm {
					strms := make([]*Strm, 0)
					for _, e := range config.Endpoints {
						strms = append(strms, fetchLocalFiles(e)...)
					}
					return strms
				}()
				for _, strm := range localstrms {
					if !strm.Check() {
						logger.Infof("[MAIN]: %s invalid, consider remove it, content: %s", strm.LocalDir+"/"+strm.Name, strm.RawURL)
						continue
					}
					logger.Debugf("[MAIN]: %s valid, content: %s", strm.LocalDir+"/"+strm.Name, strm.RawURL)
				}
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
		log.Printf("%s\n", e.Error())
	}
}
