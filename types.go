package main

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"os"

	"github.com/boltdb/bolt"
)

var (
	db *bolt.DB
)

type Strm struct {
	Name   string `json:"name"`
	Dir    string `json:"dir"`
	RawURL string `json:"raw_url"`
}

func (s Strm) Key() string {
	// 使用各算法计算key
	// md5.Sum([]byte(s.RawURL))
	// sha1.Sum([]byte(s.RawURL))
	// sha256.Sum256([]byte(s.RawURL))
	// sha512.Sum512([]byte(s.RawURL))
	//
	byts := sha1.Sum([]byte(s.RawURL))
	return fmt.Sprintf("%x", byts)
}

func (s Strm) Value() []byte {
	byts, _ := json.Marshal(s)
	return byts
}

func (s Strm) Save() error {
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

func (s Strm) GenStrm() error {
	//创建s.Dir目录
	err := os.MkdirAll(s.Dir, 0666)
	if err != nil {
		return err
	}
	//将s.RawURL写入s.Dir目录下的s.Name文件中
	return os.WriteFile(s.Dir+"/"+s.Name, []byte(s.RawURL), 0666)
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
