package redis

import (
	"bufio"
	"fmt"
	"github.com/panjf2000/ants"
	"honeypot/core/pool"
	"honeypot/utils/try"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

var kvData map[string]string

var wg sync.WaitGroup

var poolX *ants.Pool

var stop = false

func Start(addr string, done chan bool) {
	kvData = make(map[string]string)
	connDone := make(chan bool)

	// 建立socket，监听端口
	netListen, _ := net.Listen("tcp", addr)
	go closeSocket(netListen, connDone)

	defer netListen.Close()

	wg, poolX = pool.New(1)
	defer poolX.Release()

	flag := false

	for {
		wg.Add(1)
		poolX.Submit(func() {
			time.Sleep(time.Second * 2)

			conn, err := netListen.Accept()

			if err != nil {
				fmt.Printf("Redis 连接失败， error is %v\n", err)
				wg.Done()
				flag = true
				return
			}

			fmt.Printf("Redis 连接成功！")

			go handleConnection(conn, done, connDone)

			wg.Done()
		})
		if flag {
			break
		}
	}
}

func closeSocket(netListen net.Listener, connDone chan bool) {
	<-connDone
	fmt.Printf("close socket")
	if err := netListen.Close(); err != nil {
		fmt.Errorf("listen close failed error is %v\n", err)
	}
}

func closeConn(conn net.Conn, done, connDone chan bool) {
	<-done
	fmt.Printf("closeConn")
	stop = true
	conn.Close()
	connDone <- true
}

//处理 Redis 连接
func handleConnection(conn net.Conn, done, connDone chan bool) {

	go closeConn(conn, done, connDone)
	for {
		if stop {
			goto end
		}
		str := parseRESP(conn)

		switch value := str.(type) {
		case string:

			if len(value) == 0 {
				goto end
			}
			conn.Write([]byte(value))
		case []string:
			if value[0] == "SET" || value[0] == "set" {
				// 模拟 redis set

				try.Try(func() {
					key := string(value[1])
					val := string(value[2])
					kvData[key] = val

				}).Catch(func() {
					// 取不到 key 会异常
				})

				conn.Write([]byte("+OK\r\n"))
			} else if value[0] == "GET" || value[0] == "get" {
				try.Try(func() {
					// 模拟 redis get
					key := string(value[1])
					val := string(kvData[key])

					valLen := strconv.Itoa(len(val))
					str := "$" + valLen + "\r\n" + val + "\r\n"

					conn.Write([]byte(str))
				}).Catch(func() {
					conn.Write([]byte("+OK\r\n"))
				})
			} else {
				try.Try(func() {
				}).Catch(func() {
				})

				conn.Write([]byte("+OK\r\n"))
			}
			break
		default:

		}
	}
end:
	conn.Close()
}

// 解析 Redis 协议
func parseRESP(conn net.Conn) interface{} {
	if conn == nil {
		return ""
	}
	r := bufio.NewReader(conn)
	line, err := r.ReadString('\n')
	if err != nil {
		return ""
	}

	cmdType := string(line[0])
	cmdTxt := strings.Trim(string(line[1:]), "\r\n")

	switch cmdType {
	case "*":
		count, _ := strconv.Atoi(cmdTxt)
		var data []string
		for i := 0; i < count; i++ {
			line, _ := r.ReadString('\n')
			cmd_txt := strings.Trim(string(line[1:]), "\r\n")
			c, _ := strconv.Atoi(cmd_txt)
			length := c + 2
			str := ""
			for length > 0 {
				block, _ := r.Peek(length)
				if length != len(block) {

				}
				r.Discard(length)
				str += string(block)
				length -= len(block)
			}

			data = append(data, strings.Trim(str, "\r\n"))
		}
		return data
	default:
		return cmdTxt
	}
}
