package taossqlrestful

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

// TaosResq is
type TaosResq struct {
	Status string          `json:"status"`
	Head   []string        `json:"head"`
	Data   [][]interface{} `json:"data"`
	Rows   int             `json:"rows"`
	Code   int             `json:"code"`
	Desc   string          `json:"desc"`
}

// TaosConnect is
func (mc *taosConn) taosConnect(ip, user, pass, db string, port int) (taos string, err error) {
	b := []byte(fmt.Sprintf("%s:%s", user, pass))
	sqlStr := strings.NewReader("SELECT CLIENT_VERSION()")
	mc.token = base64.StdEncoding.EncodeToString(b)
	mc.reqUrl = fmt.Sprintf("http://%s:%d/rest/sqlt", ip, port)
	req, err := http.NewRequest("POST", mc.reqUrl, sqlStr)
	req.Header.Add("Authorization", "Basic "+mc.token)
	clt := http.Client{}
	resp, err := clt.Do(req)
	if err != nil {
		return "", errors.New("taos restful api fail! ")
	}
	defer resp.Body.Close()
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	taosResq := new(TaosResq)
	jsonErr := json.Unmarshal(content, &taosResq)
	if jsonErr != nil {
		return "", jsonErr
	}
	// fmt.Println(taosResq.Status)
	if taosResq.Status != "succ" {
		return "", errors.New(taosResq.Desc)
	}
	mc.db = db
	// mc.result = taosResq
	mc.dbVersion = taosResq.Data[0][0].(string)
	return "succ", nil
}

func (mc *taosConn) taosQuery(sqlstr string) (int, error) {
	mc.mu.Lock()
	// fmt.Println(sqlstr)
	clt := http.Client{}
	req1, err := http.NewRequest("POST", mc.reqUrl, strings.NewReader("use "+mc.db))
	req1.Header.Add("Authorization", "Basic "+mc.token)
	resp, err := clt.Do(req1)
	if err != nil {
		return 0, errors.New("taos restful api fail! ")
	}
	sqlStr := strings.NewReader(sqlstr)
	req, err := http.NewRequest("POST", mc.reqUrl, sqlStr)
	req.Header.Add("Authorization", "Basic "+mc.token)
	resp, err = clt.Do(req)
	if err != nil {
		return 0, errors.New("taos restful api fail! ")
	}
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}
	taosResq := new(TaosResq)
	log.Println(string(content))
	jsonErr := json.Unmarshal(content, &taosResq)
	if jsonErr != nil {
		return 0, jsonErr
	}
	if taosResq.Status != "succ" {
		return 0, errors.New(taosResq.Desc)
	}
	mc.result = taosResq
	numFields := len(taosResq.Head)
	if numFields == 1 && taosResq.Head[0] == "affected_rows" { // there are no select and show kinds of commands
		mc.affectedRows = int(taosResq.Data[0][0].(float64))
		mc.insertId = 0
		numFields = 0
	}
	defer resp.Body.Close()
	mc.mu.Unlock()
	return numFields, nil
}

func (mc *taosConn) taos_close() {
	mc.taos = ""
}
