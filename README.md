# mini-sftp-server

因为SSH端口不允许公网访问，我又需求使用`sftp`上传/下载文件（我个人不想使用FTP服务器）。  
所以我需要在Linux下部署一份`sftp-server`。  
要么我"Linux下让SSH和SFTP服务分离"（看着若干配置文件有点烦躁），  
要么我找一个可用的`sftp-server`。  

程序没找到，不过找到了一个库：  
[pkg/sftp: SFTP support for the go.crypto/ssh package](https://github.com/pkg/sftp)。  
我感觉着这个库还不错，更新还算活跃，应该会有人继续维护，  
所以我准备对它进行一下封装，搞个`sftp-server`出来。  

本程序以`https://github.com/pkg/sftp`为依赖库，  
以`https://github.com/pkg/sftp/blob/master/examples/sftp-server/main.go`为初版进行编写。  


# 使用说明

* 配置文件的说明  
参见[ConfigData.go](https://github.com/zx9229/mini-sftp-server/blob/master/ConfigData.go)顶部的注释。  


# 其他说明

* `Windows`下可以使用`freeSSHd`充当`SFTP`服务器。  
[freeSSHd and freeFTPd - open source SSH and SFTP servers for Windows](http://www.freesshd.com/)。  

* `Linux`下可以利用`OpenSSH`建立`SFTP`服务器。  
略。  
