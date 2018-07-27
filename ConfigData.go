package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"io/ioutil"

	"golang.org/x/crypto/ssh"
)

//ConfigData omit
type ConfigData struct {
	NoClientAuth  bool              //ssh.ServerConfig
	MaxAuthTries  int               //ssh.ServerConfig
	ServerVersion string            //ssh.ServerConfig
	Address       string            //例如  0.0.0.0:2222
	IsReadOnly    bool              //ReadOnly configures a Server to serve files in read-only mode.
	IsDebugMode   bool              //显示调试信息
	UserPwd       map[string]string //用户名和密码
	UserPwdMD5    map[string]string //用户名和密码的MD5值
	UseOneTimeKey bool              //使用一次性的私钥(临时密钥,在进程中临时生成,进程退出之后就丢失了)
	HostKeyFile   string            //使用指定的私钥的文件
	HostKey       string            //指定私钥的内容
	HomeDir       string            //默认目录(不填则为进程的工作目录)
}

func calcConfigData(s string, isBase64 bool) (cfg *ConfigData, err error) {
	for range "1" {
		var data []byte
		if isBase64 {
			if data, err = base64.StdEncoding.DecodeString(s); err != nil {
				break
			}
		} else {
			data = []byte(s)
		}
		cfg = new(ConfigData)
		if err = json.Unmarshal(data, cfg); err != nil {
			cfg = nil
			break
		}
	}
	return
}

func (thls *ConfigData) init() error {
	var err error

	for range "1" {

		if thls.UserPwd == nil {
			thls.UserPwd = make(map[string]string)
		}

		if thls.UserPwdMD5 == nil {
			thls.UserPwdMD5 = make(map[string]string)
		}

		if len(thls.UserPwd) == 0 && len(thls.UserPwdMD5) == 0 {
			err = errors.New("no username")
			break
		}

		if thls.UseOneTimeKey {
			thls.HostKey = calcOneTimeKey()
			if len(thls.HostKey) == 0 {
				err = errors.New("failed to calc one-time key")
				break
			}
		} else if 0 < len(thls.HostKeyFile) {
			//你可以[ssh-keygen -t rsa -f ./my_rsa]
			//然后令[HostKeyFile]的值为[./my_rsa]
			if bytes, err2 := ioutil.ReadFile(thls.HostKeyFile); err2 != nil {
				err = err2
				break
			} else {
				thls.HostKey = string(bytes)
			}
		}

		if len(thls.HostKey) == 0 {
			err = errors.New("HostKey is empty")
			break
		}
	}

	return err
}

func (thls *ConfigData) sshServerConfig() *ssh.ServerConfig {
	dstData := new(ssh.ServerConfig)
	dstData.NoClientAuth = thls.NoClientAuth
	dstData.MaxAuthTries = thls.MaxAuthTries
	dstData.ServerVersion = thls.ServerVersion
	return dstData
}

func exampleConfigData() string {
	cfgData := new(ConfigData)
	cfgData.UserPwd = make(map[string]string)
	cfgData.UserPwdMD5 = make(map[string]string)
	cfgData.Address = "localhost:2222"
	cfgData.UserPwd["root"] = "toor"
	cfgData.UserPwd["ping"] = "pong"
	cfgData.UserPwd["Scott"] = "Tiger"
	cfgData.UserPwdMD5["u"] = "83878c91171338902e0fe0fb97a8c47a" //[p]的md5值的小写.
	data, err := json.Marshal(cfgData)
	if err != nil {
		panic("UNKNOWN_ERROR")
	}
	return string(data)
}

func calcOneTimeKey() string {
	var err error
	var privateKey *rsa.PrivateKey
	if privateKey, err = rsa.GenerateKey(rand.Reader, 2048); err != nil {
		return ""
	}
	bytes := pem.EncodeToMemory(&pem.Block{
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
		Type:  "RSA PRIVATE KEY",
	})
	return string(bytes)
}
