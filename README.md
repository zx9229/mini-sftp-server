# mini-sftp-server

因为SSH端口不允许公网访问，我又需求使用`sftp`上传/下载文件（我个人对`ftp-server`有抵触情绪）。  
所以我需要在Linux下部署一份`sftp-server`。  
要么我"Linux下让SSH和SFTP服务分离"（看着若干配置文件有点烦躁），  
要么我找一个可用的`sftp-server`。  

程序没找到，不过找到了一个库：  
[pkg/sftp: SFTP support for the go.crypto/ssh package](https://github.com/pkg/sftp)。  
我感觉着这个库还不错，更新还算活跃，应该会继续有人维护，  
所以准备简单的封装一下，自己搞个`sftp-server`出来。  

本程序以`https://github.com/pkg/sftp`为依赖库，  
以`https://github.com/pkg/sftp/blob/master/examples/sftp-server/main.go`为初版进行编写。  


# 其他说明
Windows下可以用`freeSSHd`充当`sftp-server`。  
[freeSSHd and freeFTPd - open source SSH and SFTP servers for Windows](http://www.freesshd.com/)。  
