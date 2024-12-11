package main

import (
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

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
	logger.Tracef("[thread %2d]: get files from: %s, recursively: %t, include exts: %v", idx, m.CurrentRemotePath, m.IsRecursive, m.Exts)
	alistFiles, err := m.client.List(m.CurrentRemotePath, "", 1, 0, m.IsForceRefresh)
	if err != nil {
		logger.Errorf("[thread %2d]: get files from [%s] error: %s", idx, m.CurrentRemotePath, err.Error())
		return
	}
	for _, f := range alistFiles {
		if f.IsDir && m.IsRecursive {
			logger.Tracef("[thread %2d]: found sub directory [%s]", idx, f.Name)
			mm := &Mission{
				BaseURL:           m.BaseURL,
				CurrentRemotePath: m.CurrentRemotePath + "/" + f.Name,
				LocalPath: func() string {
					if m.IsCreateSubDirectory {
						return path.Join(m.LocalPath, f.Name)
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
				logger.Tracef("[thread %2d]: bind file [%s] to local dir [%s]", idx, f.Name, m.LocalPath)
				strm := Strm{
					Name: func() string {
						//change f.Name to Upper letter except the extension and return the name with extension .strm
						ext := filepath.Ext(f.Name)
						name := strings.TrimSuffix(f.Name, ext)
						//name = strings.ToUpper(name)
						//return replaceSpaceToDash(name) + ".strm"
						return name + ".strm"
					}(),
					LocalDir:  m.LocalPath,
					RemoteDir: m.CurrentRemotePath,
					RawURL:    m.BaseURL + "/d" + m.CurrentRemotePath + "/" + f.Name,
				}
				err := strm.GenStrm(true)
				if err != nil {
					logger.Errorf("[thread %2d]: save file [%s] error: %s", idx, m.CurrentRemotePath+"/"+f.Name, err.Error())
				}
				logger.Tracef("[thread %2d]: generate [%s] ==> [%s] success", idx, path.Join(strm.LocalDir, strm.Name), strm.RawURL)
				logger.Add(1)
			}
		}
	}
}

func (m *Mission) Run(concurrentNum int) {
	logger.Infof("[MAIN]: Run mission with %d threads", concurrentNum)
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

func (m *Mission) getStrm(strmChan chan *Strm) {
	threadIdx := <-m.concurrentChan
	defer func() {
		m.concurrentChan <- threadIdx
		m.wg.Done()
	}()
	logger.Debugf("[thread %2d]: get files from: %s", threadIdx, m.CurrentRemotePath)
	alistFiles, err := m.client.List(m.CurrentRemotePath, "", 1, 0, m.IsForceRefresh)
	if err != nil {
		logger.Errorf("[thread %2d]: get files from [%s] error: %s", threadIdx, m.CurrentRemotePath, err.Error())
		return
	}
	logger.Debugf("[thread %2d]: get %d files from [%s]", threadIdx, len(alistFiles), m.CurrentRemotePath)
	for _, f := range alistFiles {
		if f.IsDir && m.IsRecursive {
			logger.Debugf("[thread %2d]: found directory [%s]", threadIdx, m.CurrentRemotePath+"/"+f.Name)
			if _, ok := config.records[m.CurrentRemotePath+"/"+f.Name]; ok && config.isIncrementalUpdate {
				logger.Debugf("[thread %2d]: directory [%s] already processed and use incremental update, skip", threadIdx, m.CurrentRemotePath+"/"+f.Name)
				continue
			}
			mm := &Mission{
				BaseURL:           m.BaseURL,
				CurrentRemotePath: m.CurrentRemotePath + "/" + f.Name,
				LocalPath: func() string {
					if m.IsCreateSubDirectory {
						return path.Join(m.LocalPath, f.Name)
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
			go mm.getStrm(strmChan)
		} else if !f.IsDir {
			if checkExt(f.Name, m.Exts) {
				strm := &Strm{
					Name: func() string {
						//change f.Name to Upper letter except the extension and return the name with extension .strm
						ext := filepath.Ext(f.Name)
						name := strings.TrimSuffix(f.Name, ext)
						//name = strings.ToUpper(name)
						//return replaceSpaceToDash(name) + ".strm"
						return name + ".strm"
					}(),
					RemoteDir: m.CurrentRemotePath,
					LocalDir:  m.LocalPath,
					RawURL:    m.BaseURL + "/d" + m.CurrentRemotePath + "/" + f.Name,
					//RawURL:    m.BaseURL + "/d" + urlEncode(m.CurrentRemotePath+"/"+f.Name), //urlEncode is not necessary
				}
				strmChan <- strm
				logger.Add(1)
			}
		}
	}
}

// This function returns a slice of pointers to Strm objects
func (m *Mission) GetAllStrm(concurrentNum int) []*Strm {
	// Log the number of concurrent threads
	logger.Infof("[MAIN]: Run mission with %d threads", concurrentNum)
	// Create a channel for concurrent threads
	m.concurrentChan = make(chan int, concurrentNum)
	// Push the threads to the channel
	for i := 0; i < concurrentNum; i++ {

		logger.Debugf("[MAIN]: Push thread %d to concurrent channel", i)
		m.concurrentChan <- i
	}
	// Create a waitgroup
	m.wg = &sync.WaitGroup{}
	// Add one to the waitgroup
	m.wg.Add(1)
	// Create a channel for strm objects
	strmChan := make(chan *Strm, 1000)
	// Run the getStrm function in a goroutine
	go m.getStrm(strmChan)
	// Create a channel to stop the goroutine
	stopChan := make(chan struct{})
	// Create a channel to return the results
	resultChan := make(chan []*Strm, 1)
	// Run the goroutine to collect the strm objects
	go func() {
		// Create an empty slice of strm pointers
		strms := make([]*Strm, 0)
		// Loop indefinitely
		for {
			// Select from the stop channel or the strm channel
			select {
			// If the stop channel is closed, return the results
			case <-stopChan:
				resultChan <- strms
			// If a strm object is received, append it to the slice
			case strm := <-strmChan:
				strms = append(strms, strm)
			}
			// Sleep for 5 milliseconds
			time.Sleep(5 * time.Millisecond)
		}
	}()
	// Wait for the waitgroup to finish
	m.wg.Wait()
	// Send a stop signal to the goroutine
	stopChan <- struct{}{}
	// Return the results
	return <-resultChan
}
