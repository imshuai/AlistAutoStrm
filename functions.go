package main

import (
	"net/url"
	"path/filepath"
	"strings"
	"sync"

	sdk "github.com/imshuai/alistsdk-go"
)

func FetchRemoteFile(wg *sync.WaitGroup, c *sdk.Client, localDirectory, path string, isCreateSubDirectory, isRecursive bool, exts []string, files chan<- File) {
	defer wg.Done()
	wg.Add(1)
	ff, err := GetFiles(c, path, isRecursive, config.Exts)
	if err != nil {
		logger.Errorf("get files from [%s] error: %s", path, err.Error())
		return
	}
	for _, f := range ff {
		f.LocalDir = func() string {
			if isCreateSubDirectory {
				return strings.ReplaceAll(f.RemoteDir, path, localDirectory)
			} else {
				return localDirectory
			}
		}()
		logger.Debugf("bind file [%s] to local dir [%s]", f.Name, f.LocalDir)
		files <- f
	}
}

func GetFiles(c *sdk.Client, path string, isRecursive bool, exts []string) ([]File, error) {
	logger.Debugf("GetFiles from: %s, recursively: %t, include exts: %v", path, isRecursive, exts)
	files := make([]File, 0)
	aFiles, err := c.List(path, "", 1, 0, true)
	if err != nil {
		return nil, err
	}
	if !isRecursive {
		for _, v := range aFiles {
			f := File{
				v,
				path,
				"",
			}
			if checkExt(v.Name, exts) {
				files = append(files, f)
			}
		}
		return files, nil
	}
	for _, v := range aFiles {
		if v.IsDir {
			t, e := GetFiles(c, path+"/"+v.Name, true, exts)
			if e != nil {
				return nil, e
			}
			if len(t) > 0 {
				for _, tt := range t {
					if checkExt(tt.Name, exts) {
						files = append(files, tt)
					}
				}
			}
		} else {
			f := File{
				v,
				path,
				"",
			}
			if checkExt(v.Name, exts) {
				files = append(files, f)
			}
		}
	}
	return files, nil
}

func goGetFiles(wg *sync.WaitGroup, c *sdk.Client, localDirectory, remoteRootPath, entryPath string, isCreateSubDirectory, isRecursive bool, isForceRefresh bool, exts []string, files chan<- File, concurrentChan chan struct{}) {
	defer func() {
		wg.Done()
		concurrentChan <- struct{}{}
	}()
	<-concurrentChan

	logger.Debugf("GetFiles from: %s, recursively: %t, include exts: %v", entryPath, isRecursive, exts)
	aFiles, err := c.List(entryPath, "", 1, 0, isForceRefresh)
	if err != nil {
		logger.Errorf("get files from [%s] error: %s", entryPath, err.Error())
		return
	}
	for _, v := range aFiles {
		if v.IsDir && isRecursive {
			wg.Add(1)
			go goGetFiles(wg, c, localDirectory, remoteRootPath, entryPath+"/"+v.Name, isCreateSubDirectory, isRecursive, isForceRefresh, exts, files, concurrentChan)
		} else {
			f := File{
				v,
				entryPath,
				"",
			}
			if checkExt(v.Name, exts) {
				f.LocalDir = func() string {
					if isCreateSubDirectory {
						return strings.ReplaceAll(f.RemoteDir, remoteRootPath, localDirectory)
					} else {
						return localDirectory
					}
				}()
				logger.Debugf("bind file [%s] to local dir [%s]", f.Name, f.LocalDir)
				files <- f
			}
		}
	}
}

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

// 删除字符串中所有不是0-9a-zA-Z以及'-'和'@'的字符
func removeSpecialChars(s string) string {
	return strings.Map(func(r rune) rune {
		if (r >= '0' && r <= '9') || (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || r == '-' || r == '@' {
			return r
		}
		return -1
	}, s)
}
