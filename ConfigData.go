package main

import (
	"encoding/base64"
	"encoding/json"

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
	UserPwds      map[string]string //用户名密码
	HostKey       string            //如果没有它,会 failed to handshake (ssh: server has no host keys)
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

func (thls *ConfigData) sshServerConfig() *ssh.ServerConfig {
	dstData := new(ssh.ServerConfig)
	dstData.NoClientAuth = thls.NoClientAuth
	dstData.MaxAuthTries = thls.MaxAuthTries
	dstData.ServerVersion = thls.ServerVersion
	return dstData
}

func exampleConfigData() string {
	exampleCfg := new(ConfigData)
	exampleCfg.UserPwds = make(map[string]string)
	exampleCfg.Address = "localhost:2222"
	exampleCfg.UserPwds["root"] = "toor"
	exampleCfg.UserPwds["ping"] = "pong"
	exampleCfg.UserPwds["Scott"] = "Tiger"
	data, err := json.Marshal(exampleCfg)
	if err != nil {
		panic("UNKNOWN_ERROR")
	}
	return string(data)
}
