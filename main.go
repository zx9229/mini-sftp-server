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
	GlobalDebugStream     = ioutil.Discard
	GlobalConfigData      *ConfigData
	GlobalSshServerConfig *ssh.ServerConfig
	HardCodedHostKey      = `-----BEGIN RSA PRIVATE KEY-----
MIIEoAIBAAKCAQEAkiUqO6yn+UKgQmUvrnv92xsDx8wteFlJeHG5GQUK5fx2gOO2
ZVomv/+ko2lNBtlVNAnjn1jlBRZnKSWs9gpIi2gcjx3ppWp8Ck4m7Eu7vxa/PPU9
EAQ9o0uxGgJE+62wZqPKr3ufr1oCokCWqPBqKW4Qh8Mvhb5Sr9ZGvtygftS0jUMA
25U/qsYV1udoxRkLeD13SRppXPNJLTMTIjSTdvEdKmZfmKM4GGicnkS4JvrR9KpM
CSlOkz6NCW0UBMg45zQ3Kl8qLu5XG4ibsdaMwCnc/ASJzrwUvo0XZ00eo1wNEdHx
RIU6vqtdD1f9lLotYNAaYB6VbkX6lzL+mj1bfwIBIwKCAQA2SFGD4QswsljINDZI
H2zq+2fN3h+EeO9nQC7Ox1vRxCwD/M637kjoOmG5CdrIB5Ss7bryCxM8Z2glOeEo
L7SLjRHsA8vPudZM+HTbbJYxCHLqwX0Uk9xhOV8JqRJO2hzy7GE53XXTapNDlFU3
bz1f2GyKMo3+eeQyrqyQCM3l9qwPjJ1QhO0gncsjAeS/wYkRuxooZCdssSK+hrCs
Ts/OWsvHUU8gjJKuUmEk+e57A5IVlp5WUDAzj4k93X6NUn1wexaCMM5A/M9mK0kJ
4RKCZ8M7w2XPBF0RES7neUUZfsm/QEif5/wpWng2fywKIRjGCCngzxiDo1BerPeR
MstzAoGBAMGYvaUPuHr15HSMyuxRoS4qsFBW4DoiRwueCRPEvfxsr44AFyNdD4SR
IsSt2s4G3MkGgkxg65vhaMVfTYyqoTKNWxzjyeI1eMoNLvP7XbYYqpWhPycHcR4l
nBFDxa6LGTVKpfbP1eA5WsuSf/Xc8RmQKeV38kcTP/dtPmrf2xWLAoGBAMFAz7RB
2Vn6m0NOUncMvZOnqRuZLIoCWS7JSTdmLEAtdH45Ed1aGTxPe760Pyz8T3+gV2kO
oOInBJE+o9AscFIrA92s/jBeL16sgqf+VfOXLNdDt1Ctoa1Dm2B5QPEiY6Trpkqe
yJ/umBLCvLiI0m6ny04BEm7ROqDv+4MMFJhdAoGAY5BhiBa2paMHxulSaugnAczP
tEnvqN5trjQEqxS5eoEJ1AAL5k0d7Gfl/r/PnSgZxnhgRImdveGjmLSriivd33vl
txYQDe+dNLZSqVyz2f4OlhhpnwskO2PMmyorJpCt4OSP3gR8nzNwhfOSQ+34Vkok
LN6Z22j8U1zBA8OVPkcCgYAxsZR+zxqiG97IKhU0jj9ge5HihnkqzWdjzVv4TXkX
0SyVfGOt8pjGXZTZRErCbMP8PyxreMpIyDRf3OhLeSQyYtUbvsUFH4iGDxpIdJnC
S3HuNfvwLKXquZz7jOTQSqvolF317lDYqxEpZUZ4mDYcdEo4oTCgJyxVRQYpAxs9
HwKBgAE/T4oHr2zspJcPCXCphZbwpx8eue4MhqVk72D+VfNm99GfITN9te8WxxhL
puEW8ZV4pJZMKuNpICcixh9CeVUoK+W5wUczmre3+HWoBpXTkNu1Nd1EXMFPMnuK
q1YwWb4VHBqECkkpUsyhtB3t7QycFciEDKdDIujQDHI7dXp4
-----END RSA PRIVATE KEY-----`

//对于HostKey,你可以[ssh-keygen -t rsa -f ./tmprsa]然后将[./tmprsa]文件的内容拷贝到这里来.
)

// Based on example server code from golang.org/x/crypto/ssh and server_standalone
func main() {

	var (
		isHelp   bool
		confName string
	)

	flag.BoolVar(&isHelp, "help", false, "show this help")
	flag.StringVar(&confName, "conf", "", "config filename")
	flag.Parse()

	if isHelp {
		flag.Usage()
		fmt.Println()
		fmt.Println()
		fmt.Println(exampleConfigData())
		return
	}

	if 0 < len(confName) {
		if bytes, err := ioutil.ReadFile(confName); err != nil {
			log.Fatal("Failed to load conf", err)
		} else {
			if GlobalConfigData, err = calcConfigData(string(bytes), false); err != nil {
				log.Fatal("Failed to load conf", err)
			}
		}
	}

	if GlobalConfigData.IsDebugMode {
		GlobalDebugStream = os.Stderr
	}

	if len(GlobalConfigData.HostKey) == 0 {
		GlobalConfigData.HostKey = HardCodedHostKey
	}

	hostKey, err := ssh.ParsePrivateKey([]byte(GlobalConfigData.HostKey))
	if err != nil {
		log.Printf("ParsePrivateKey, err=%v", err)
		os.Exit(100)
	}

	// An SSH server is represented by a ServerConfig, which holds
	// certificate details and handles authentication of ServerConns.
	GlobalSshServerConfig = GlobalConfigData.sshServerConfig()
	GlobalSshServerConfig.PasswordCallback = tmpPasswordCallback
	GlobalSshServerConfig.AddHostKey(hostKey)

	// Once a ServerConfig has been configured, connections can be
	// accepted.
	listener, err := net.Listen("tcp", GlobalConfigData.Address)
	if err != nil {
		log.Fatal("failed to listen for connection", err)
	}
	fmt.Printf("Listening on %v\n", listener.Addr())

	for {
		nConn, err := listener.Accept()
		if err != nil {
			log.Fatal("failed to accept incoming connection", err)
		}

		// Before use, a handshake must be performed on the incoming
		// net.Conn.
		_, chans, reqs, err := ssh.NewServerConn(nConn, GlobalSshServerConfig)
		if err != nil {
			log.Fatal("failed to handshake", err)
		}
		fmt.Fprintf(GlobalDebugStream, "SSH server established\n")

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
		fmt.Fprintf(GlobalDebugStream, "Incoming channel: %s\n", newChannel.ChannelType())
		if newChannel.ChannelType() != "session" {
			newChannel.Reject(ssh.UnknownChannelType, "unknown channel type")
			fmt.Fprintf(GlobalDebugStream, "Unknown channel type: %s\n", newChannel.ChannelType())
			continue
		}
		channel, requests, err := newChannel.Accept()
		if err != nil {
			log.Fatal("could not accept channel.", err)
		}
		fmt.Fprintf(GlobalDebugStream, "Channel accepted\n")

		// Sessions have out-of-band requests such as "shell",
		// "pty-req" and "env".  Here we handle only the
		// "subsystem" request.
		go func(in <-chan *ssh.Request) {
			for req := range in {
				fmt.Fprintf(GlobalDebugStream, "Request: %v\n", req.Type)
				ok := false
				switch req.Type {
				case "subsystem":
					fmt.Fprintf(GlobalDebugStream, "Subsystem: %s\n", req.Payload[4:])
					if string(req.Payload[4:]) == "sftp" {
						ok = true
					}
				}
				fmt.Fprintf(GlobalDebugStream, " - accepted: %v\n", ok)
				req.Reply(ok, nil)
			}
		}(requests)

		serverOptions := []sftp.ServerOption{
			sftp.WithDebug(GlobalDebugStream),
		}

		if GlobalConfigData.IsReadOnly {
			serverOptions = append(serverOptions, sftp.ReadOnly())
			fmt.Fprintf(GlobalDebugStream, "Read-only server\n")
		} else {
			fmt.Fprintf(GlobalDebugStream, "Read write server\n")
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
	fmt.Fprintf(GlobalDebugStream, "Trying to auth user "+remoteConn.User()+"\n")

	for range "1" {
		if GlobalConfigData.UserPwds == nil {
			err = errors.New("User does not exist")
			log.Println(err)
			break
		}
		curPwd, isOk := GlobalConfigData.UserPwds[remoteConn.User()]
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
