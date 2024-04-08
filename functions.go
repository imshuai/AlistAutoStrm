package main

import (
	"net/url"
	"path/filepath"
	"strings"
)

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

// 替换字符串中的空格为'-'
func replaceSpaceToDash(s string) string {
	return strings.ReplaceAll(s, " ", "-")
}
