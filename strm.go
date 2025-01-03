package main

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path"

	"github.com/boltdb/bolt"
)

type Strm struct {
	Name      string `json:"name"`
	LocalDir  string `json:"local_dir"`
	RemoteDir string `json:"remote_dir"`
	RawURL    string `json:"raw_url"`
}

// 生成Strm对象的唯一键
func (s *Strm) Key() string {
	byts := sha1.Sum([]byte(s.RawURL))
	return fmt.Sprintf("%x", byts)
}

// 将Strm对象序列化为JSON字节数组
func (s *Strm) Value() []byte {
	byts, _ := json.Marshal(s)
	return byts
}

// 删除Strm对象
func (s *Strm) Delete() error {
	err := os.RemoveAll(path.Join(s.LocalDir, s.Name))
	if err != nil {
		return err
	}
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("strm"))
		if b == nil {
			return fmt.Errorf("bucket not found")
		}
		return b.Delete([]byte(s.Key()))
	})
}

// 保存Strm对象
func (s *Strm) Save() error {
	return db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("strm"))
		if err != nil {
			return err
		}
		return b.Put([]byte(s.Key()), s.Value())
	})
}

// 生成Strm文件
func (s *Strm) GenStrm(overwrite bool) error {
	err := os.MkdirAll(s.LocalDir, 0666)
	if err != nil {
		return err
	}
	_, err = os.Stat(path.Join(s.LocalDir, s.Name))
	if !overwrite && !os.IsNotExist(err) {
		return fmt.Errorf("file %s already exists and overwrite is false", path.Join(s.LocalDir, s.Name))
	}
	return os.WriteFile(path.Join(s.LocalDir, s.Name), []byte(s.RawURL), 0666)
}

// 检查Strm文件是否有效
func (s *Strm) Check() bool {
	logger.Infof("Checking %s", s.LocalDir+"/"+s.Name)
	resp, err := http.Head(s.RawURL)
	if err != nil {
		logger.Errorf("http.Head(%s) error: %v", s.RawURL, err)
		return false
	}
	defer resp.Body.Close()
	if resp.StatusCode == 302 || (resp.StatusCode == 200 && (resp.Header.Get("Content-Type") == "video/mp4" || resp.Header.Get("Content-Type") == "application/octet-stream")) {
		return true
	}
	return false
}

// 根据rawUrl获取Strm对象
func GetStrm(rawUrl string) (*Strm, error) {
	var strm Strm
	strm.RawURL = rawUrl
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("strm"))
		if b == nil {
			return fmt.Errorf("bucket not found")
		}
		v := b.Get([]byte(strm.Key()))
		if v == nil {
			return fmt.Errorf("key not found")
		}
		return json.Unmarshal(v, &strm)
	})
	return &strm, err
}

// 获取记录集合
func GetRecordCollection() (map[string]int, error) {
	var records map[string]int
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("strm"))
		byts := b.Get([]byte("records"))
		if byts == nil {
			return fmt.Errorf("records not found, must use update-database first")
		}
		return json.Unmarshal(byts, &records)
	})
	if err != nil {
		return nil, err
	}
	return records, nil
}

// 保存记录集合
func SaveRecordCollection(records map[string]int) error {
	return db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("strm"))
		if err != nil {
			return err
		}
		byts, err := json.Marshal(records)
		if err != nil {
			return err
		}
		return b.Put([]byte("records"), byts)
	})
}
