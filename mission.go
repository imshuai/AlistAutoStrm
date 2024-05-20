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
	Exts                 []string
	IsCreateSubDirectory bool
	IsRecursive          bool
	IsForceRefresh       bool
	client               *sdk.Client
	wg                   *sync.WaitGroup
	concurrentChan       chan struct{}
}

// Walk walks the current remote path
func (m *Mission) walk() {
	m.concurrentChan <- struct{}{}
	defer func() {
		<-m.concurrentChan
		m.wg.Done()
	}()
	logger.Debugf("GetFiles from: %s, recursively: %t, include exts: %v", m.CurrentRemotePath, m.IsRecursive, m.Exts)
	alistFiles, err := m.client.List(m.CurrentRemotePath, "", 1, 0, m.IsForceRefresh)
	if err != nil {
		logger.Errorf("get files from [%s] error: %s", m.CurrentRemotePath, err.Error())
	}
	for _, f := range alistFiles {
		if f.IsDir && m.IsRecursive {
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
				logger.Debugf("bind file [%s] to local dir [%s]", f.Name, m.LocalPath)
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
					RawURL: config.Endpoint + "/d" + urlEncode(m.CurrentRemotePath+"/"+f.Name),
				}
				err := strm.GenStrm()
				if err != nil {
					logger.Errorf("save file [%s] error: %s", m.CurrentRemotePath+"/"+f.Name, err.Error())
				}
				logger.Infof("generate [%s] ==> [%s] success", strm.Dir+"/"+strm.Name, strm.RawURL)
			}
		}
	}
}

func (m *Mission) Run(concurrentNum int) {
	m.concurrentChan = make(chan struct{}, concurrentNum)
	m.wg = &sync.WaitGroup{}
	m.wg.Add(1)
	go m.walk()
	m.wg.Wait()
}
