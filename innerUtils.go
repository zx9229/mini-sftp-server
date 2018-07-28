package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

func loadConfigContent(isStdin bool, base64Data string, filename string, isOffset bool) (content string, isBase64 bool, err error) {
	content = ""
	isBase64 = false

	if isStdin {
		//其实,可以从标准输入中读取整个配置文件的,
		//因为文件的内容可能有多行,不太好读取,
		//所以仅支持从标准输入中读取配置文件的base64编码后的内容.
		if _, err = fmt.Scanln(&content); err != nil {
			content = ""
		}
		isBase64 = true
		return
	}

	if 0 < len(base64Data) {
		content = base64Data
		isBase64 = true
		return
	}

	if 0 < len(filename) {
		if isOffset && !path.IsAbs(filename) {
			filename = path.Join(os.Args[0][:strings.LastIndexAny(os.Args[0], `/\`)+1], filename)
		}
		var bytes []byte
		if bytes, err = ioutil.ReadFile(filename); err == nil {
			content = string(bytes)
		}
		isBase64 = false
		return
	}

	err = errors.New("unable to load the config content")
	return
}
