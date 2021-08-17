package tdengine

import (
	"fmt"
	"testing"
)

func TestInsertPotData(t *testing.T) {
	//Start()
	err := InsertPotData("redis", "")
	if err != nil {
		fmt.Errorf("insert pot data failed, error is %v", err)
	}
}

func TestGetPotData(t *testing.T) {
	err := GetPotData()
	if err != nil {
		fmt.Errorf("get pot data failed, error is %v", err)
	}
}
