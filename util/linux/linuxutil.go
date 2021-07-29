package linux

import (
	"fmt"
	"os/exec"
)

func CheckPort(port string) bool {
	checkStatement := fmt.Sprintf("lsof -i:%s ", port)
	output, _ := exec.Command("sh", "-c", checkStatement).CombinedOutput()
	if len(output) > 0 {
		return true
	}
	return false
}
