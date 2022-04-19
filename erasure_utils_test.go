package grasure

import (
	"fmt"
	"testing"
)

func TestGetMyIP(t *testing.T){
	ret := GetMyIP()
	fmt.Println(ret)
}
