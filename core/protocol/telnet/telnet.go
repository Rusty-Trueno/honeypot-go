package telnet

import (
	"bufio"
	"fmt"
	"github.com/bitly/go-simplejson"
	"github.com/panjf2000/ants"
	"honeypot/core/pool"
	"honeypot/util/file"
	"honeypot/util/json"
	"net"
	"strings"
	"sync"
	"time"
)

var wg sync.WaitGroup

var poolX *ants.Pool

// 服务端连接
func Start(address string, done chan bool) {
	l, err := net.Listen("tcp", address)

	if err != nil {
		fmt.Println(err.Error())
		done <- true
	}

	defer l.Close()

	wg, poolX = pool.New(10)
	defer poolX.Release()

	go closeSocket(l, poolX, done)

	flag := false

	for {
		wg.Add(1)
		poolX.Submit(func() {
			time.Sleep(time.Second * 2)

			conn, err := l.Accept()

			if err != nil {
				fmt.Printf("Telnet", "127.0.0.1", "Telnet 连接失败", err)
				wg.Done()
				flag = true
				return
			}

			arr := strings.Split(conn.RemoteAddr().String(), ":")

			var id string

			fmt.Printf("Telnet", arr[0], "已经连接")

			// 根据连接开启会话, 这个过程需要并行执行
			go handleSession(conn, id, done)

			wg.Done()
		})
	}
}

func getJson() *simplejson.Json {
	res, err := json.GetTelnet()

	if err != nil {
		fmt.Printf("HFish", "127.0.0.1", "解析 Telnet JSON 文件失败", err)
	}
	return res
}

// 会话处理
func handleSession(conn net.Conn, id string, done chan bool) {
	fmt.Println("Session started")
	reader := bufio.NewReader(conn)

	for {
		str, err := reader.ReadString('\n')

		// telnet命令
		if err == nil {
			str = strings.TrimSpace(str)

			if !processTelnetCommand(str, done) {
				conn.Close()
				break
			}

			res := getJson()

			fileName := res.Get("command").Get(str).MustString()

			if fileName == "" {
				fileName = res.Get("command").Get("default").MustString()
			}

			output := file.ReadLibsText("telnet", fileName)

			conn.Write([]byte(output + "\r\n"))
		} else {
			// 发生错误
			fmt.Println("Session closed")
			conn.Close()
			break
		}
	}
}

// telent协议命令
func processTelnetCommand(str string, done chan bool) bool {
	// @close指令表示终止本次会话
	if strings.HasPrefix(str, "@close") {
		fmt.Println("Session closed")
		// 告知外部需要断开连接
		return false
		// @shutdown指令表示终止服务进程
	} else if strings.HasPrefix(str, "@shutdown") {
		fmt.Println("Server shutdown")
		// 往通道中写入0, 阻塞等待接收方处理
		done <- true
		return false
	}
	return true
}

func closeSocket(netListen net.Listener, poolX *ants.Pool, done chan bool) {
	<-done
	fmt.Printf("close socket\n")
	if err := netListen.Close(); err != nil {
		fmt.Errorf("listen close failed error is %v\n", err)
	}
	poolX.Release()
}
