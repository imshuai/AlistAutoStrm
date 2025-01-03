package main

import (
	"io"
	"net/http"
	"os"
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
	AltExts              []string
	IsCreateSubDirectory bool
	IsRecursive          bool
	IsForceRefresh       bool
	client               *sdk.Client
	wg                   *sync.WaitGroup
	concurrentChan       chan int
}

func (m *Mission) getStrm(strmChan chan *Strm) {
	threadIdx := <-m.concurrentChan
	defer func() {
		m.concurrentChan <- threadIdx
		m.wg.Done()
	}()
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
				AltExts:              m.AltExts,
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
			} else if checkExt(f.Name, m.AltExts) {
				// check if the file is in the altExts list
				// if it is, download the file to the current local directory
				logger.Debugf("[thread %2d]: found file [%s], download to [%s]", threadIdx, m.CurrentRemotePath+"/"+f.Name, m.LocalPath)
				// 检查文件是否已存在
				filePath := path.Join(m.LocalPath, f.Name)
				if _, statErr := os.Stat(filePath); statErr == nil {
					logger.Debugf("[thread %2d]: file [%s] already exists, skip download", threadIdx, filePath)
					continue
				}

				// 下载文件
				req, err := http.NewRequest("GET", f.RawURL, nil)
				if err != nil {
					logger.Errorf("[thread %2d]: create request for [%s] error: %s", threadIdx, f.RawURL, err.Error())
					continue
				}

				// 设置常见的浏览器User-Agent
				req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

				resp, err := http.DefaultClient.Do(req)
				if err != nil {
					logger.Errorf("[thread %2d]: download [%s] error: %s", threadIdx, f.RawURL, err.Error())
					continue
				}
				defer resp.Body.Close()

				// 创建目标文件
				out, err := os.Create(filePath)
				if err != nil {
					logger.Errorf("[thread %2d]: create file [%s] error: %s", threadIdx, filePath, err.Error())
					continue
				}
				defer out.Close()

				// 写入文件
				_, err = io.Copy(out, resp.Body)
				if err != nil {
					logger.Errorf("[thread %2d]: write file [%s] error: %s", threadIdx, filePath, err.Error())
					continue
				}

				// 检查文件大小是否匹配
				fileInfo, err := os.Stat(filePath)
				if err != nil {
					logger.Errorf("[thread %2d]: get file info [%s] error: %s", threadIdx, filePath, err.Error())
					os.Remove(filePath)
					continue
				}

				if fileInfo.Size() != f.Size {
					logger.Errorf("[thread %2d]: file size mismatch for [%s], expected %d but got %d",
						threadIdx, filePath, f.Size, fileInfo.Size())
					os.Remove(filePath)
					continue
				}

				logger.Debugf("[thread %2d]: successfully downloaded [%s] to [%s], size %d bytes",
					threadIdx, f.RawURL, filePath, fileInfo.Size())

			}
		}
	}
}

// 这个函数返回一个指向 Strm 对象的指针切片
func (m *Mission) GetAllStrm(concurrentNum int) []*Strm {
	// 记录并发线程的数量
	logger.Infof("[MAIN]: Run mission with %d threads", concurrentNum)
	// 创建一个用于并发线程的通道
	m.concurrentChan = make(chan int, concurrentNum)
	// 将线程推送到通道中
	for i := 0; i < concurrentNum; i++ {
		logger.Debugf("[MAIN]: Push thread %d to concurrent channel", i)
		m.concurrentChan <- i
	}
	// 创建一个等待组
	m.wg = &sync.WaitGroup{}
	// 向等待组添加一个计数
	m.wg.Add(1)
	// 创建一个用于 strm 对象的通道
	strmChan := make(chan *Strm, 1000)
	// 在一个 goroutine 中运行 getStrm 函数
	go m.getStrm(strmChan)
	// 创建一个用于停止 goroutine 的通道
	stopChan := make(chan struct{})
	// 创建一个用于返回结果的通道
	resultChan := make(chan []*Strm, 1)
	// 运行 goroutine 来收集 strm 对象
	go func() {
		// 创建一个空的 strm 指针切片
		strms := make([]*Strm, 0)
		// 无限循环
		for {
			// 从停止通道或 strm 通道中选择
			select {
			// 如果停止通道关闭，返回结果
			case <-stopChan:
				resultChan <- strms
			// 如果接收到一个 strm 对象，将其追加到切片中
			case strm := <-strmChan:
				strms = append(strms, strm)
			}
			// 休眠 5 毫秒
			time.Sleep(5 * time.Millisecond)
		}
	}()
	// 等待等待组完成
	m.wg.Wait()
	// 发送停止信号给 goroutine
	stopChan <- struct{}{}
	// 返回结果
	return <-resultChan
}
