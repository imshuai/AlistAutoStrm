package main

import (
	"path/filepath"
	"strings"
	"sync"

	sdk "github.com/imshuai/alistsdk-go"
)

// Mission is a struct that holds the mission data
type Mission struct {
	CurrentRemotePath    string
	LocalPath            string
	BaseURL              string
	Exts                 []string
	IsCreateSubDirectory bool
	IsRecursive          bool
	IsForceRefresh       bool
	client               *sdk.Client
	wg                   *sync.WaitGroup
	concurrentChan       chan int
}

// Walk walks the current remote path
func (m *Mission) walk() {
	idx := <-m.concurrentChan
	defer func() {
		m.concurrentChan <- idx
		m.wg.Done()
	}()
	logger.Debugf("[thread %2d]: get files from: %s, recursively: %t, include exts: %v", idx, m.CurrentRemotePath, m.IsRecursive, m.Exts)
	alistFiles, err := m.client.List(m.CurrentRemotePath, "", 1, 0, m.IsForceRefresh)
	if err != nil {
		logger.Errorf("[thread %2d]: get files from [%s] error: %s", idx, m.CurrentRemotePath, err.Error())
		return
	}
	for _, f := range alistFiles {
		if f.IsDir && m.IsRecursive {
			logger.Debugf("[thread %2d]: found sub directory [%s]", idx, f.Name)
			mm := &Mission{
				CurrentRemotePath: m.CurrentRemotePath + "/" + f.Name,
				LocalPath: func() string {
					if m.IsCreateSubDirectory {
						return m.LocalPath + "/" + f.Name
					} else {
						return m.LocalPath
					}
				}(),
				Exts:                 m.Exts,
				IsCreateSubDirectory: m.IsCreateSubDirectory,
				IsRecursive:          m.IsRecursive,
				IsForceRefresh:       m.IsForceRefresh,
				client:               m.client,
				wg:                   m.wg,
				concurrentChan:       m.concurrentChan,
			}
			m.wg.Add(1)
			go mm.walk()
		} else if !f.IsDir {
			if checkExt(f.Name, m.Exts) {
				logger.Debugf("[thread %2d]: bind file [%s] to local dir [%s]", idx, f.Name, m.LocalPath)
				strm := Strm{
					Name: func() string {
						//change f.Name to Upper letter except the extension and return the name with extension .strm
						ext := filepath.Ext(f.Name)
						name := strings.TrimSuffix(f.Name, ext)
						//name = strings.ToUpper(name)
						//return replaceSpaceToDash(name) + ".strm"
						return name + ".strm"
					}(),
					Dir:    m.LocalPath,
					RawURL: m.BaseURL + "/d" + urlEncode(m.CurrentRemotePath+"/"+f.Name),
				}
				err := strm.GenStrm()
				if err != nil {
					logger.Errorf("[thread %2d]: save file [%s] error: %s", idx, m.CurrentRemotePath+"/"+f.Name, err.Error())
				}
				logger.Debugf("[thread %2d]: generate [%s] ==> [%s] success", idx, strm.Dir+"/"+strm.Name, strm.RawURL)
			}
		}
	}
}

func (m *Mission) Run(concurrentNum int) {
	logger.Infof("[MAIN]: Run mission with concurrent number: %d", concurrentNum)
	m.concurrentChan = make(chan int, concurrentNum)
	for i := 0; i < concurrentNum; i++ {
		logger.Debugf("[MAIN]: Push thread %d to concurrent channel", i)
		m.concurrentChan <- i
	}
	m.wg = &sync.WaitGroup{}
	m.wg.Add(1)
	go m.walk()
	logger.Info("[MAIN]: Wait for walk to finish")
	m.wg.Wait()
	logger.Info("[MAIN]: All threads finished")
}
