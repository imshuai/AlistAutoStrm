package main

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"os"
	"path"

	"github.com/boltdb/bolt"
)

var (
	db *bolt.DB
)

type Strm struct {
	Name      string `json:"name"`
	LocalDir  string `json:"local_dir"`
	RemoteDir string `json:"remote_dir"`
	RawURL    string `json:"raw_url"`
}

func (s *Strm) Key() string {
	// 使用各算法计算key
	// md5.Sum([]byte(s.RawURL))
	// sha1.Sum([]byte(s.RawURL))
	// sha256.Sum256([]byte(s.RawURL))
	// sha512.Sum512([]byte(s.RawURL))
	//
	byts := sha1.Sum([]byte(s.RawURL))
	return fmt.Sprintf("%x", byts)
}

func (s *Strm) Value() []byte {
	byts, _ := json.Marshal(s)
	return byts
}

func (s *Strm) Delete() error {
	//TODO 使用boltdb实现Strm对象的删除逻辑
	err := os.RemoveAll(path.Join(s.LocalDir, s.Name))
	if err != nil {
		return err
	}
	// return db.Update(func(tx *bolt.Tx) error {
	// 	// 获取名为strm的bucket
	// 	b := tx.Bucket([]byte("strm"))
	// 	if b == nil {
	// 		// 如果bucket不存在，返回错误
	// 		return fmt.Errorf("bucket not found")
	// 	}
	// 	// 根据key删除value
	// 	return b.Delete([]byte(s.Key()))
	// })
	return nil
}

func (s *Strm) Save() error {
	//TODO 使用boltdb实现Strm对象的保存逻辑
	return db.Update(func(tx *bolt.Tx) error {
		//创建一个名为"strm"的bucket
		b, err := tx.CreateBucketIfNotExists([]byte("strm"))
		if err != nil {
			return err
		}
		//将Strm对象的key和value存入bucket中
		return b.Put([]byte(s.Key()), s.Value())
	})
}

func (s *Strm) GenStrm(overwrite bool) error {
	//创建s.Dir目录
	err := os.MkdirAll(s.LocalDir, 0666)
	if err != nil {
		return err
	}
	// 如果s.Dir目录下已经存在s.Name文件，并且overwrite为false，则返回错误
	_, err = os.Stat(path.Join(s.LocalDir, s.Name))
	if !overwrite && !os.IsNotExist(err) {
		return fmt.Errorf("file %s already exists and overwrite is false", path.Join(s.LocalDir, s.Name))
	}
	// 将s.RawURL写入s.Dir目录下的s.Name文件中
	return os.WriteFile(path.Join(s.LocalDir, s.Name), []byte(s.RawURL), 0666)
}

func GetStrm(rawUrl string) (*Strm, error) {
	//TODO 使用boltdb实现根据key获取Strm对象的逻辑
	var strm Strm

	strm.RawURL = rawUrl

	err := db.View(func(tx *bolt.Tx) error {
		// 获取名为strm的bucket
		b := tx.Bucket([]byte("strm"))
		if b == nil {
			// 如果bucket不存在，返回错误
			return fmt.Errorf("bucket not found")
		}
		// 根据key获取value
		v := b.Get([]byte(strm.Key()))
		if v == nil {
			// 如果key不存在，返回错误
			return fmt.Errorf("key not found")
		}
		// 将value反序列化为Strm对象
		return json.Unmarshal(v, &strm)
	})
	// 返回Strm对象和错误
	return &strm, err
}

func GetRecordCollection() (map[string]int, error) {
	var records map[string]int
	err := db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte("strm"))
		byts := b.Get([]byte("records"))
		return json.Unmarshal(byts, &records)
	})
	if err != nil {
		return nil, err
	}
	return records, nil
}

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
