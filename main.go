// An example SFTP server implementation using the golang SSH package.
// Serves the whole filesystem visible to the user, and has a hard-coded username and password,
// so not for real use!
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

var (
	globalDebugStream     = ioutil.Discard
	globalConfigData      *ConfigData
	globalSSHServerConfig *ssh.ServerConfig
)

// Based on example server code from golang.org/x/crypto/ssh and server_standalone
func main() {

	var (
		isHelp     bool
		isStdin    bool
		base64Data string
		confName   string
	)

	flag.BoolVar(&isHelp, "help", false, "show this help")
	flag.BoolVar(&isStdin, "stdin", false, "read base64 encoding data from stdin")
	flag.StringVar(&base64Data, "base64", "", "base64 encoding of the config file")
	flag.StringVar(&confName, "conf", "", "config filename")
	flag.Parse()

	if isHelp {
		flag.Usage()
		fmt.Println()
		fmt.Println(exampleConfigData())
		return
	}

	content, isBase64, err := loadConfigContent(isStdin, base64Data, confName)
	if err != nil {
		log.Fatalln("loadConfigContent,", err)
	}

	if globalConfigData, err = calcConfigData(content, isBase64); err != nil {
		log.Fatalln("calcConfigData,", err)
	}

	if err := globalConfigData.init(); err != nil {
		log.Fatalln("init config,", err)
	}

	if globalConfigData.IsDebugMode {
		globalDebugStream = os.Stderr
	}

	var hostKey ssh.Signer
	if hostKey, err = ssh.ParsePrivateKey([]byte(globalConfigData.HostKey)); err != nil {
		log.Fatalln("ParsePrivateKey", err)
	}

	// An SSH server is represented by a ServerConfig, which holds
	// certificate details and handles authentication of ServerConns.
	globalSSHServerConfig = globalConfigData.sshServerConfig()
	globalSSHServerConfig.PasswordCallback = tmpPasswordCallback
	globalSSHServerConfig.AddHostKey(hostKey)

	// Once a ServerConfig has been configured, connections can be
	// accepted.
	listener, err := net.Listen("tcp", globalConfigData.Address)
	if err != nil {
		log.Fatal("failed to listen for connection", err)
	}
	log.Printf("Listening on %v", listener.Addr())

	for {
		nConn, err := listener.Accept()
		if err != nil {
			log.Println("failed to accept incoming connection", err)
			continue
		}

		// Before use, a handshake must be performed on the incoming
		// net.Conn.
		_, chans, reqs, err := ssh.NewServerConn(nConn, globalSSHServerConfig)
		if err != nil {
			log.Println("failed to handshake", err)
			nConn.Close()
			continue
		}
		fmt.Fprintf(globalDebugStream, "SSH server established\n")

		// The incoming Request channel must be serviced.
		go ssh.DiscardRequests(reqs)

		go handleChannels(chans)
	}
}

func handleChannels(chans <-chan ssh.NewChannel) {
	// Service the incoming Channel channel.
	for newChannel := range chans {
		// Channels have a type, depending on the application level
		// protocol intended. In the case of an SFTP session, this is "subsystem"
		// with a payload string of "<length=4>sftp"
		fmt.Fprintf(globalDebugStream, "Incoming channel: %s\n", newChannel.ChannelType())
		if newChannel.ChannelType() != "session" {
			newChannel.Reject(ssh.UnknownChannelType, "unknown channel type")
			fmt.Fprintf(globalDebugStream, "Unknown channel type: %s\n", newChannel.ChannelType())
			continue
		}
		channel, requests, err := newChannel.Accept()
		if err != nil {
			log.Fatal("could not accept channel.", err)
		}
		fmt.Fprintf(globalDebugStream, "Channel accepted\n")

		// Sessions have out-of-band requests such as "shell",
		// "pty-req" and "env".  Here we handle only the
		// "subsystem" request.
		go func(in <-chan *ssh.Request) {
			for req := range in {
				fmt.Fprintf(globalDebugStream, "Request: %v\n", req.Type)
				ok := false
				switch req.Type {
				case "subsystem":
					fmt.Fprintf(globalDebugStream, "Subsystem: %s\n", req.Payload[4:])
					if string(req.Payload[4:]) == "sftp" {
						ok = true
					}
				}
				fmt.Fprintf(globalDebugStream, " - accepted: %v\n", ok)
				req.Reply(ok, nil)
			}
		}(requests)

		serverOptions := []sftp.ServerOption{
			sftp.WithDebug(globalDebugStream),
		}

		if globalConfigData.IsReadOnly {
			serverOptions = append(serverOptions, sftp.ReadOnly())
			fmt.Fprintf(globalDebugStream, "Read-only server\n")
		} else {
			fmt.Fprintf(globalDebugStream, "Read write server\n")
		}

		server, err := sftp.NewServer(
			channel,
			serverOptions...,
		)
		if err != nil {
			log.Fatal(err)
		}
		if err := server.Serve(); err == io.EOF {
			server.Close()
			log.Print("sftp client exited session.")
		} else if err != nil {
			log.Fatal("sftp server completed with error:", err)
		}
	}
}

func tmpPasswordCallback(remoteConn ssh.ConnMetadata, password []byte) (p *ssh.Permissions, err error) {
	fmt.Fprintf(globalDebugStream, "Trying to auth user "+remoteConn.User()+"\n")

	for range "1" {
		if globalConfigData.UserPwds == nil {
			err = errors.New("User does not exist")
			log.Println(err)
			break
		}
		curPwd, isOk := globalConfigData.UserPwds[remoteConn.User()]
		if !isOk {
			err = errors.New("User does not exist")
			log.Println(err)
			break
		}
		if curPwd != string(password) {
			err = errors.New("Incorrect password")
			log.Println(err)
			break
		}
	}

	return
}

func loadConfigContent(isStdin bool, base64Data string, filename string) (content string, isBase64 bool, err error) {
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
