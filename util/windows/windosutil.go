package windows

import (
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

func CheckPort(port string) bool {
	res := -1
	var outBytes bytes.Buffer
	cmdStr := fmt.Sprintf("netstat -ano -p tcp | findstr %s", port)
	cmd := exec.Command("cmd", "/c", cmdStr)
	cmd.Stdout = &outBytes
	cmd.Run()
	resStr := outBytes.String()
	r := regexp.MustCompile(`\s\d+\s`).FindAllString(resStr, -1)
	if len(r) > 0 {
		pid, err := strconv.Atoi(strings.TrimSpace(r[0]))
		if err != nil {
			res = -1
		} else {
			res = pid
		}
	}
	return res != -1
}
