package json

import (
	"fmt"
	"github.com/bitly/go-simplejson"
	"io/ioutil"
)

var telnetJson []byte

func init() {
	file, err := ioutil.ReadFile("./libs/telnet/config.json")

	if err != nil {
		fmt.Printf("HFish", "127.0.0.1", "读取文件失败", err)
	}

	telnetJson = file
}

func GetTelnet() (*simplejson.Json, error) {
	res, err := simplejson.NewJson(telnetJson)
	return res, err
}
