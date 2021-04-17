package ssh

import (
	"fmt"
	"github.com/bitly/go-simplejson"
	"golang.org/x/crypto/ssh/terminal"
	ssh "honeypot/core/protocol/ssh/gliderlabs"
	"honeypot/util/file"
	"honeypot/util/json"
	"io"
	"strings"
)

var clientData map[string]string

func getJson() *simplejson.Json {
	res, err := json.GetSsh()

	if err != nil {
		fmt.Printf("HFish", "127.0.0.1", "解析 SSH JSON 文件失败", err)
	}
	return res
}

func Start(addr string, done chan bool) {
	clientData = make(map[string]string)

	ssh.ListenAndServe(
		addr,
		func(s ssh.Session) {
			res := getJson()

			term := terminal.NewTerminal(s, res.Get("hostname").MustString())
			for {
				line, rerr := term.ReadLine()

				if rerr != nil {
					break
				}

				if line == "exit" {
					break
				}

				fileName := res.Get("command").Get(line).MustString()

				output := file.ReadLibsText("ssh", fileName)

				//id := clientData[s.RemoteAddr().String()]

				io.WriteString(s, output+"\n")
			}
		},
		ssh.PasswordAuth(func(s ssh.Context, password string) bool {
			//info := s.User() + "&&" + password

			arr := strings.Split(s.RemoteAddr().String(), ":")

			fmt.Printf("SSH", arr[0], "已经连接")

			var id string

			sshStatus := "2" //conf.Get("ssh", "status")

			if sshStatus == "2" {
				// 高交互模式
				res := getJson()
				accountx := res.Get("account")
				passwordx := res.Get("password")

				if accountx.MustString() == s.User() && passwordx.MustString() == password {
					clientData[s.RemoteAddr().String()] = id
					return true
				}
			}

			// 低交互模式，返回账号密码不正确
			return false
		}),
	)
}
