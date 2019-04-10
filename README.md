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

* 下载源码  
`go get -u -v github.com/zx9229/mini-sftp-server`

* 查看帮助  
`.\mini-sftp-server.exe -help`

* 使用示例  
首先，`.\mini-sftp-server.exe -help > cfg.json`  
然后，修改`cfg.json`文件。  
最后，`.\mini-sftp-server.exe -conf cfg.json`  
亦或者`.\mini-sftp-server.exe -conf cfg.json -force`

* 配置文件的说明  
参见[ConfigData.go](https://github.com/zx9229/mini-sftp-server/blob/master/ConfigData.go)顶部的注释。  


# 其他说明

* 我想(在Windows/Linux下)格式化`JSON`格式的配置文件  
[stedolan/jq: Command-line JSON processor](https://github.com/stedolan/jq)。  
下载：`wget https://github.com/stedolan/jq/releases/download/jq-1.5/jq-linux32`。  
使用：`./jq-linux32 . cfg.json > cfg.json`。  

* `Windows`下可以使用`freeSSHd`充当`SFTP`服务器。  
[freeSSHd and freeFTPd - open source SSH and SFTP servers for Windows](http://www.freesshd.com/)。  

* `Linux`下可以利用`OpenSSH`建立`SFTP`服务器。  
略。  

* `Windows`下设置开机自启动程序(不依赖账户登录)  
思路：创建一个任务计划，它以系统账户(`SYSTEM`)运行，它在开机时(`ONSTART`)运行。无需账户登录。  
命令：`SCHTASKS /Create /TN 任务名 /RU SYSTEM /SC ONSTART /TR 程序的路径和文件名`。  
示例：`SCHTASKS /Create /TN MSFTPS /RU SYSTEM /SC ONSTART /TR "C:\exeFolder\msftps.exe -conf cfg.json"`。  
解释：创建一个任务计划，名为`MSFTPS`，运行命令是`C:\exeFolder\msftps.exe -conf cfg.json`。  
注意：如果需要更细致的设置，请在`任务计划程序`里面进行设置。  
注意：测试的时候，请关闭防火墙！

* `Linux`下设置开机自启动程序  
略。


# 备注说明  

`Windows`下可以使用`WinSCP`连接该服务程序。  
`Linux`下可以使用`sshpass`为`sftp`填入密码：`sshpass -p 密码 sftp -P 端口     用户名@主机:远程文件名 本地文件名`。  
`Linux`下可以使用`sshpass`为`sftp`填入密码：`sshpass -p 密码 sftp -oPort=端口 用户名@主机:远程文件名 本地文件名`。  
```
sshpass -p 密码 sftp -oPort=端口 用户@主机IP地址  <<-EOF
一些命令
EOF
sshpass -p 密码 sftp -oPort=端口 用户@192.168.2.247  <<-EOF
get read_me.txt
EOF
```
